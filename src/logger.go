package main

/*
 logger.go: syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/ipfix"
	"github.com/tehmaze/netflow/netflow5"
	"github.com/tehmaze/netflow/read"
	"github.com/tehmaze/netflow/session"

	gosnmp "github.com/soniah/gosnmp"

	"go.etcd.io/bbolt"
	"gopkg.in/mcuadros/go-syslog.v2"
	syslogv2 "gopkg.in/mcuadros/go-syslog.v2"
	astilog "github.com/asticode/go-astilog"
)

var (
	logCh = make(chan logEnt, 100)
)

func logger(ctx context.Context) {
	astilog.Debug("start logger")
	var syslogdRunning = false
	var trapdRunning = false
	var netflowdRunning = false
	var stopSyslogd chan bool
	var stopTrapd  chan bool
	var stopNetflowd chan bool
	logBuffer := []logEnt{}
	for {
		select {
		case <-ctx.Done():
			{
				saveLogBuffer(logBuffer)
				if syslogdRunning {
					syslogdRunning = false
					close(stopSyslogd)
				}
				if netflowdRunning {
					netflowdRunning = false
					close(stopNetflowd)
				}
				if trapdRunning {
					trapdRunning = false
					close(stopTrapd)
				}
				astilog.Debug("Stop logger")
				return
			}
		case l := <-logCh:
			{
				logBuffer = append(logBuffer, l)
			}
		case <-time.Tick(time.Second * 1):
			{
				saveLogBuffer(logBuffer)
				if mapConf.EnableSyslogd && !syslogdRunning {
					stopSyslogd = make(chan bool)
					syslogdRunning = true
					go syslogd(stopSyslogd)
					astilog.Debug("start syslogd")
				} else if !mapConf.EnableSyslogd && syslogdRunning {
					close(stopSyslogd) 
					syslogdRunning = false
					astilog.Debug("stop syslogd")
				}
				if mapConf.EnableTrapd && !trapdRunning {
					stopTrapd = make(chan bool)
					trapdRunning = true
					go trapd(stopTrapd)
					astilog.Debug("start trapd")
				} else if !mapConf.EnableTrapd && trapdRunning {
					close(stopTrapd)
					astilog.Debug("stop trapd")
				}
				if mapConf.EnableNetflowd && !netflowdRunning {
					stopNetflowd = make(chan bool)
					netflowdRunning = true
					go netflowd(stopNetflowd)
					astilog.Debug("start netflowd")
				} else if !mapConf.EnableNetflowd && netflowdRunning {
					close(stopNetflowd)
					astilog.Debug("stop netflowd")
				}
			}
		}
	}
}

func saveLogBuffer(logBuffer []logEnt) {
	if len(logBuffer) < 1 {
		return
	}
	if db == nil {
		return
	}
	db.Batch(func(tx *bbolt.Tx) error {
		syslog := tx.Bucket([]byte("syslog"))
		netflow5 := tx.Bucket([]byte("netflow5"))
		ipfix := tx.Bucket([]byte("ipfix"))
		trap := tx.Bucket([]byte("trap"))
		for _, l := range logBuffer {
			k := fmt.Sprintf("%016x", l.Time)
			switch l.Type {
			case "syslog":
				syslog.Put([]byte(k), []byte(l.Log))
			case "netflow5":
				netflow5.Put([]byte(k), []byte(l.Log))
			case "ipfix":
				ipfix.Put([]byte(k), []byte(l.Log))
			case "trap":
				trap.Put([]byte(k), []byte(l.Log))
			}
		}
		return nil
	})
}

func syslogd(stopCh chan bool) {
	syslogCh := make(syslog.LogPartsChannel)
	server := syslogv2.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(syslog.NewChannelHandler(syslogCh))
	server.ListenUDP("0.0.0.0:514")
	server.ListenTCP("0.0.0.0:514")
	server.Boot()

	for {
		select {
		case <-stopCh:
			{
				server.Kill()
				return
			}
		case l := <-syslogCh:
			{
				s, err := json.Marshal(l)
				if err == nil {
					logCh <- logEnt{
						Time: time.Now().UnixNano(),
						Type: "syslog",
						Log:  string(s),
					}
				}
			}
		}
	}
}

func netflowd(stopCh chan bool) {
	var readSize = 2 << 16
	var addr *net.UDPAddr
	var err error
	if addr, err = net.ResolveUDPAddr("udp", ":2055"); err != nil {
		log.Fatal(err)
	}
	var server *net.UDPConn
	if server, err = net.ListenUDP("udp", addr); err != nil {
		log.Fatal(err)
	}
	defer server.Close()
	if err = server.SetReadBuffer(readSize); err != nil {
		log.Fatal(err)
	}
	decoders := make(map[string]*netflow.Decoder)
	for {
		select {
		case <-stopCh:
			{
				return
			}
		default:
			{
				server.SetReadDeadline(time.Now().Add(time.Second * 1))
				buf := make([]byte, 8192)
				var remote *net.UDPAddr
				var octets int
				if octets, remote, err = server.ReadFromUDP(buf); err != nil {
					if !strings.Contains(err.Error(), "timeout") {
						astilog.Debug(err)
					}
					continue
				}
				d, found := decoders[remote.String()]
				if !found {
					s := session.New()
					d = netflow.NewDecoder(s)
					decoders[remote.String()] = d
				}
				m, err := d.Read(bytes.NewBuffer(buf[:octets]))
				if err != nil {
					astilog.Debug("decoder error:", err)
					continue
				}
				switch p := m.(type) {
				case *netflow5.Packet:
					{
						logNetflow5(p)
					}
				case *ipfix.Message:
					{
						logIPFIX(p)
					}
				}
			}
		}
	}
}

func logIPFIX(p *ipfix.Message) {
	l := logEnt{
		Time: time.Now().UnixNano(),
		Type: "ipfix",
		Log:  "",
	}
	for _, ds := range p.DataSets {
		if ds.Records == nil {
			continue
		}
		for _, dr := range ds.Records {
			var record = make(map[string]interface{})
			for _, f := range dr.Fields {
				if f.Translated != nil {
					if f.Translated.Name != "" {
						record[f.Translated.Name] = f.Translated.Value
					} else {
						record[fmt.Sprintf("%d.%d", f.Translated.EnterpriseNumber, f.Translated.InformationElementID)] = f.Bytes
					}
				} else {
					record["raw"] = f.Bytes
				}
			}
			s, err := json.Marshal(record)
			if err != nil {
				fmt.Println(err)
			}
			l.Log = string(s)
			logCh <- l
		}
	}
}

func logNetflow5(p *netflow5.Packet) {
	l := logEnt{
		Time: time.Now().UnixNano(),
		Type: "netflow5",
		Log:  "",
	}
	var record = make(map[string]interface{})
	record["header"] = p.Header.String()
	for _, r := range p.Records {
		record["srcAddr"] = r.SrcAddr
		record["srcPort"] = r.SrcPort
		record["dstAddr"] = r.DstAddr
		record["dstPort"] = r.DstPort
		record["nextHop"] = r.NextHop
		record["bytes"] = r.Bytes
		record["packets"] = r.Packets
		record["first"] = r.First
		record["last"] = r.Last
		record["tcpflags"] = r.TCPFlags
		record["tcpflagsStr"] = read.TCPFlags(r.TCPFlags)
		record["protocol"] = r.Protocol
		record["protocolStr"] = read.Protocol(r.Protocol)
		record["tos"] = r.ToS
		record["srcAs"] = r.SrcAS
		record["dstAs"] = r.DstAS
		record["srcMask"] = r.SrcMask
		record["dstMask"] = r.DstMask
		s, err := json.Marshal(record)
		if err != nil {
			fmt.Println(err)
		}
		l.Log = string(s)
		logCh <- l
	}
}

func trapd(stopCh chan bool) {
	tl := gosnmp.NewTrapListener()
	tl.OnNewTrap = func(s *gosnmp.SnmpPacket, u *net.UDPAddr) {
		l := logEnt{
			Time: time.Now().UnixNano(),
			Type: "trap",
			Log:  "",
		}
		var record = make(map[string]interface{})
		record["FromAddress"] = u.String()
		record["Timestamp"] = s.Timestamp
		record["Timestamp"] = s.Enterprise
		record["Timestamp"] = s.GenericTrap
		record["Timestamp"] = s.SpecificTrap
		for _, vb := range s.Variables {
			key := mib.OIDToName(vb.Name)
			switch vb.Type {
			case gosnmp.ObjectIdentifier:
				record[key] = mib.OIDToName(vb.Value.(string))
			case gosnmp.OctetString:
				record[key] = vb.Value.(string)
			default:
				record[key] = gosnmp.ToBigInt(vb.Value).Int64()
			}
		}
		js, err := json.Marshal(record)
		if err != nil {
			astilog.Debug(err)
		}
		l.Log = string(js)
		logCh <- l
	}
	defer tl.Close()
	go func() {
		tl.Listen("0.0.0.0:162")
		astilog.Debug("Trap Listen End")
	}()
	for {
		select {
		case <- stopCh:
			{
				astilog.Debug("Trap Listen Done")
				return
			}
		}
	}
}
