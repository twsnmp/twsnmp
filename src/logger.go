package main

/*
 logger.go: syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/tehmaze/netflow"
	"github.com/tehmaze/netflow/ipfix"
	"github.com/tehmaze/netflow/netflow5"
	"github.com/tehmaze/netflow/read"
	"github.com/tehmaze/netflow/session"

	gosnmp "github.com/soniah/gosnmp"

	astilog "github.com/asticode/go-astilog"
	"go.etcd.io/bbolt"
	"gopkg.in/mcuadros/go-syslog.v2"
	syslogv2 "gopkg.in/mcuadros/go-syslog.v2"
)

var (
	logCh = make(chan *logEnt, 100)
)

func logger(ctx context.Context) {
	astilog.Debug("start logger")
	var syslogdRunning = false
	var trapdRunning = false
	var netflowdRunning = false
	var stopSyslogd chan bool
	var stopTrapd chan bool
	var stopNetflowd chan bool
	timer := time.NewTicker(time.Second * 10)
	logBuffer := []*logEnt{}
	for {
		select {
		case <-ctx.Done():
			{
				timer.Stop()
				if len(logBuffer) > 0 {
					saveLogBuffer(logBuffer)
					logBuffer = []*logEnt{}
				}
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
		case <-timer.C:
			{
				if len(logBuffer) > 0 {
					astilog.Infof("Save Logs %d", len(logBuffer))
					saveLogBuffer(logBuffer)
					logBuffer = []*logEnt{}
				}
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

func saveLogBuffer(logBuffer []*logEnt) {
	if db == nil {
		astilog.Errorf("saveLogBuffer DB Not open")
		return
	}
	db.Batch(func(tx *bbolt.Tx) error {
		syslog := tx.Bucket([]byte("syslog"))
		netflow := tx.Bucket([]byte("netflow"))
		ipfix := tx.Bucket([]byte("ipfix"))
		trap := tx.Bucket([]byte("trap"))
		arplog := tx.Bucket([]byte("arplog"))
		for _, l := range logBuffer {
			k := fmt.Sprintf("%016x", l.Time)
			s, err := json.Marshal(l)
			if err != nil {
				return err
			}
			switch l.Type {
			case "syslog":
				syslog.Put([]byte(k), []byte(s))
			case "netflow":
				netflow.Put([]byte(k), []byte(s))
			case "ipfix":
				ipfix.Put([]byte(k), []byte(s))
			case "trap":
				trap.Put([]byte(k), []byte(s))
			case "arplog":
				arplog.Put([]byte(k), []byte(s))
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
	astilog.Debug("syslogd start")
	for {
		select {
		case <-stopCh:
			{
				astilog.Debug("syslogd stop")
				server.Kill()
				return
			}
		case l := <-syslogCh:
			{
				s, err := json.Marshal(l)
				if err == nil {
					logCh <- &logEnt{
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
	astilog.Debug("netflowd start")
	if addr, err = net.ResolveUDPAddr("udp", ":2055"); err != nil {
		astilog.Errorf("netflowd err=%v", err)
		return
	}
	var server *net.UDPConn
	if server, err = net.ListenUDP("udp", addr); err != nil {
		astilog.Errorf("netflowd err=%v", err)
		return
	}
	defer server.Close()
	if err = server.SetReadBuffer(readSize); err != nil {
		astilog.Errorf("netflowd err=%v", err)
		return
	}
	decoders := make(map[string]*netflow.Decoder)
	buf := make([]byte, 8192)
	for {
		select {
		case <-stopCh:
			{
				astilog.Debug("netflowd stop")
				return
			}
		default:
			{
				server.SetReadDeadline(time.Now().Add(time.Second * 2))
				var remote *net.UDPAddr
				var octets int
				if octets, remote, err = server.ReadFromUDP(buf); err != nil {
					if !strings.Contains(err.Error(), "timeout") {
						astilog.Errorf("netflowd err=%v", err)
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
					astilog.Errorf("netflowd err=%v", err)
					continue
				}
				switch p := m.(type) {
				case *netflow5.Packet:
					{
						logNetflow(p)
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
				astilog.Errorf("logIPFIX err=%v", err)
				continue
			}
			logCh <- &logEnt{
				Time: time.Now().UnixNano(),
				Type: "ipfix",
				Log:  string(s),
			}
			if _, ok := record["sourceIPv4Address"]; ok {
				flowReportCh <- &flowReportEnt{
					Time:    time.Now().UnixNano(),
					SrcIP:   record["sourceIPv4Address"].(net.IP).String(),
					SrcPort: int(record["sourceTransportPort"].(uint16)),
					DstIP:   record["destinationIPv4Address"].(net.IP).String(),
					DstPort: int(record["destinationTransportPort"].(uint16)),
					Prot:    int(record["protocolIdentifier"].(uint8)),
					Bytes:   int64(record["octetDeltaCount"].(uint64)),
				}
			}
		}
	}
}

func logNetflow(p *netflow5.Packet) {
	var record = make(map[string]interface{})
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
		logCh <- &logEnt{
			Time: time.Now().UnixNano(),
			Type: "netflow",
			Log:  string(s),
		}
		flowReportCh <- &flowReportEnt{
			Time:    time.Now().UnixNano(),
			SrcIP:   record["srcAddr"].(net.IP).String(),
			SrcPort: int(record["srcPort"].(uint16)),
			DstIP:   record["dstAddr"].(net.IP).String(),
			DstPort: int(record["dstPort"].(uint16)),
			Prot:    int(record["protocol"].(uint8)),
			Bytes:   int64(r.Bytes),
		}

	}
}

func trapd(stopCh chan bool) {
	tl := gosnmp.NewTrapListener()
	tl.OnNewTrap = func(s *gosnmp.SnmpPacket, u *net.UDPAddr) {
		var record = make(map[string]interface{})
		record["FromAddress"] = u.String()
		record["Timestamp"] = s.Timestamp
		record["Enterprise"] = s.Enterprise
		record["GenericTrap"] = s.GenericTrap
		record["SpecificTrap"] = s.SpecificTrap
		record["Variables"] = ""
		vbs := ""
		for _, vb := range s.Variables {
			key := mib.OIDToName(vb.Name)
			val := ""
			switch vb.Type {
			case gosnmp.ObjectIdentifier:
				val = mib.OIDToName(vb.Value.(string))
			case gosnmp.OctetString:
				val = vb.Value.(string)
			default:
				val = fmt.Sprintf("%d", gosnmp.ToBigInt(vb.Value).Int64())
			}
			vbs += fmt.Sprintf("%s=%s\n", key, val)
		}
		record["Variables"] = vbs
		js, err := json.Marshal(record)
		if err != nil {
			astilog.Debug(err)
		}
		logCh <- &logEnt{
			Time: time.Now().UnixNano(),
			Type: "trap",
			Log:  string(js),
		}
	}
	defer tl.Close()
	go func() {
		tl.Listen("0.0.0.0:162")
		astilog.Debug("Trap Listen End")
	}()
	for {
		select {
		case <-stopCh:
			{
				astilog.Debug("Trap Listen Done")
				return
			}
		}
	}
}
