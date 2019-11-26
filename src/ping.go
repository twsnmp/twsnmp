package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"net"
	"runtime"
	"sync"
	"syscall"
	"time"

	astilog "github.com/asticode/go-astilog"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

const (
	timeSliceLength = 8
	trackerLength   = 8
	protocolICMP    = 1
)

type pingStat int

const (
	pingStart = iota
	pingOK
	pingTimuout
	pingNoRoute
	pingOtherError
)

type pingEnt struct {
	Target   string
	Router   string
	Timeout  int
	Retry    int
	Size     int
	ipaddr   *net.IPAddr
	id       int
	sequence int
	Tracker  int64
	Stat     pingStat
	Time     int64
	done     chan bool
}

type packet struct {
	bytes  []byte
	nbytes int
	ttl    int
}

var pingSendCh = make(chan *pingEnt, 10)

type safePingMap struct {
	v   map[int64]*pingEnt
	mux sync.Mutex
}

var pingMap = &safePingMap{
	v: make(map[int64]*pingEnt),
}

func (m *safePingMap) set(k int64, v *pingEnt) bool {
	m.mux.Lock()
	if _, ok := m.v[k]; ok {
		m.mux.Unlock()
		return false
	}
	m.v[k] = v
	m.mux.Unlock()
	return true
}

func (m *safePingMap) get(k int64) *pingEnt {
	m.mux.Lock()
	defer m.mux.Unlock()
	if r, ok := m.v[k]; ok {
		return r
	}
	return nil
}

func (m *safePingMap) del(k int64) {
	m.mux.Lock()
	delete(m.v, k)
	m.mux.Unlock()
}

var randGen = rand.New(rand.NewSource(time.Now().UnixNano()))

func doPing(ip string, timeout, retry, size int) *pingEnt {
	var err error
	var p = &pingEnt{
		Target:  ip,
		Timeout: timeout,
		Retry:   retry,
		Size:    size,
		id:      randGen.Intn(math.MaxInt16),
		Tracker: randGen.Int63n(math.MaxInt64),
		done:    make(chan bool),
	}
	if p.ipaddr, err = net.ResolveIPAddr("ip", ip); err != nil {
		p.Stat = pingOtherError
		return p
	}
	// 念のためTrackerの重複を防ぐ
	for !pingMap.set(p.Tracker, p) {
		astilog.Debugf("Dup Tracker %v len=%d", p, len(pingMap.v))
		p.Tracker++
	}
	defer func() {
		pingMap.del(p.Tracker)
	}()
	for i := 0; i < p.Retry+1; i++ {
		pingSendCh <- p
		if p.waitPingResp() {
			return p
		}
	}
	astilog.Debugf("Ping timeout retry over %s", ip)
	p.Stat = pingTimuout
	return p
}

func (p *pingEnt) waitPingResp() bool {
	select {
	case <-p.done:
		return true
	case <-time.After(time.Duration(p.Timeout) * time.Second):
		astilog.Debugf("Ping Timeout %v", p)
		return false
	}
}

func (p *pingEnt) sendICMP(conn *icmp.PacketConn) error {
	var dst net.Addr = p.ipaddr
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		dst = &net.UDPAddr{IP: p.ipaddr.IP, Zone: p.ipaddr.Zone}
	}
	t := append(timeToBytes(time.Now()), intToBytes(p.Tracker)...)
	if remainSize := p.Size - timeSliceLength - trackerLength; remainSize > 0 {
		t = append(t, bytes.Repeat([]byte{1}, remainSize)...)
	}

	body := &icmp.Echo{
		ID:   p.id,
		Seq:  p.sequence,
		Data: t,
	}

	msg := &icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: body,
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return err
	}
	for {
		if _, err := conn.WriteTo(msgBytes, dst); err != nil {
			if neterr, ok := err.(*net.OpError); ok {
				if neterr.Err == syscall.ENOBUFS {
					continue
				}
			}
			return err
		}
		break
	}
	return nil
}

func pingBackend(ctx context.Context) {
	netProto := "udp4"
	if runtime.GOOS == "windows" {
		netProto = "ip4:icmp"
	}
	conn, err := icmp.ListenPacket(netProto, "0.0.0.0")
	if err != nil {
		astilog.Fatalf("pingBackend err=%v", err)
	}
	defer conn.Close()
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-pingSendCh:
			if err := p.sendICMP(conn); err != nil {
				astilog.Debugf("sendICMP err=%v", err)
			}
		default:
			bytes := make([]byte, 2048)
			conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			var n, ttl int
			var err error
			var cm *ipv4.ControlMessage
			n, cm, _, err = conn.IPv4PacketConn().ReadFrom(bytes)
			if cm != nil {
				ttl = cm.TTL
			}
			if err != nil {
				if neterr, ok := err.(*net.OpError); ok {
					if neterr.Timeout() {
						// Read timeout
						continue
					}
				}
				astilog.Errorf("pingBackend err=%v", err)
				continue
			}
			if err := processPacket(&packet{bytes: bytes, nbytes: n, ttl: ttl}); err != nil {
				astilog.Debugf("pingBackend processPacket err=%v", err)
			}
		}
	}
}

func processPacket(recv *packet) error {
	receivedAt := time.Now()
	var m *icmp.Message
	var err error
	if m, err = icmp.ParseMessage(protocolICMP, recv.bytes); err != nil {
		return fmt.Errorf("error parsing icmp message: %s", err.Error())
	}
	if m.Type != ipv4.ICMPTypeEchoReply {
		astilog.Debugf("icmp message type != ICMPTypeEchoReply  : %v", m)
		return nil
	}
	switch pkt := m.Body.(type) {
	case *icmp.Echo:
		if len(pkt.Data) < timeSliceLength+trackerLength {
			return fmt.Errorf("insufficient data received; got: %d %v", len(pkt.Data), pkt.Data)
		}
		tracker := bytesToInt(pkt.Data[timeSliceLength:])
		timestamp := bytesToTime(pkt.Data[:timeSliceLength])
		if p := pingMap.get(tracker); p != nil {
			p.Time = receivedAt.Sub(timestamp).Nanoseconds()
			p.Stat = pingOK
			p.done <- true
		}
	default:
		// Very bad, not sure how this can happen
		return fmt.Errorf("invalid ICMP echo reply; type: '%T', '%v'", pkt, pkt)
	}
	return nil
}

func bytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}

func bytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}

func intToBytes(tracker int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(tracker))
	return b
}
