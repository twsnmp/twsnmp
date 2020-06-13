package main

/*
 logger.go: syslog,tarp,netflow5,ipfixをログに記録する
*/

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/json"
	"io/ioutil"

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

	"go.etcd.io/bbolt"
	"gopkg.in/mcuadros/go-syslog.v2"
	syslogv2 "gopkg.in/mcuadros/go-syslog.v2"
)

var (
	logCh = make(chan *logEnt, 100)
)

func logger(ctx context.Context) {
	astiLogger.Debug("start logger")
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
				astiLogger.Debug("Stop logger")
				return
			}
		case l := <-logCh:
			{
				logBuffer = append(logBuffer, l)
			}
		case <-timer.C:
			{
				if len(logBuffer) > 0 {
					astiLogger.Infof("Save Logs %d", len(logBuffer))
					saveLogBuffer(logBuffer)
					astiLogger.Infof("logSize=%d compLogSize=%d", logSize, compLogSize)
					logBuffer = []*logEnt{}
				}
				if mapConf.EnableSyslogd && !syslogdRunning {
					stopSyslogd = make(chan bool)
					syslogdRunning = true
					go syslogd(stopSyslogd)
					astiLogger.Debug("start syslogd")
				} else if !mapConf.EnableSyslogd && syslogdRunning {
					close(stopSyslogd)
					syslogdRunning = false
					astiLogger.Debug("stop syslogd")
				}
				if mapConf.EnableTrapd && !trapdRunning {
					stopTrapd = make(chan bool)
					trapdRunning = true
					go trapd(stopTrapd)
					astiLogger.Debug("start trapd")
				} else if !mapConf.EnableTrapd && trapdRunning {
					close(stopTrapd)
					trapdRunning = false
					astiLogger.Debug("stop trapd")
				}
				if mapConf.EnableNetflowd && !netflowdRunning {
					stopNetflowd = make(chan bool)
					netflowdRunning = true
					go netflowd(stopNetflowd)
					astiLogger.Debug("start netflowd")
				} else if !mapConf.EnableNetflowd && netflowdRunning {
					close(stopNetflowd)
					netflowdRunning = false
					astiLogger.Debug("stop netflowd")
				}
			}
		}
	}
}

var logSize = 0
var compLogSize = 0

func compressLog(s []byte) []byte {
	var b bytes.Buffer
	f, _ := flate.NewWriter(&b, flate.DefaultCompression)
	if _, err := f.Write(s); err != nil {
		return s
	}
	if err := f.Flush(); err != nil {
		return s
	}
	if err := f.Close(); err != nil {
		return s
	}
	return b.Bytes()
}

func deCompressLog(s []byte) []byte {
	r := flate.NewReader(bytes.NewBuffer(s))
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return s
	}
	return d
}

func saveLogBuffer(logBuffer []*logEnt) {
	if db == nil {
		astiLogger.Errorf("saveLogBuffer DB Not open")
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
			logSize += len(s)
			if len(s) > 100 {
				s = compressLog(s)
			}
			compLogSize += len(s)
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
	astiLogger.Debug("syslogd start")
	for {
		select {
		case <-stopCh:
			{
				astiLogger.Debug("syslogd stop")
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
	astiLogger.Debug("netflowd start")
	if addr, err = net.ResolveUDPAddr("udp", ":2055"); err != nil {
		astiLogger.Errorf("netflowd err=%v", err)
		return
	}
	var server *net.UDPConn
	if server, err = net.ListenUDP("udp", addr); err != nil {
		astiLogger.Errorf("netflowd err=%v", err)
		return
	}
	defer server.Close()
	if err = server.SetReadBuffer(readSize); err != nil {
		astiLogger.Errorf("netflowd err=%v", err)
		return
	}
	decoders := make(map[string]*netflow.Decoder)
	buf := make([]byte, 8192)
	for {
		select {
		case <-stopCh:
			{
				astiLogger.Debug("netflowd stop")
				return
			}
		default:
			{
				server.SetReadDeadline(time.Now().Add(time.Second * 2))
				var remote *net.UDPAddr
				var octets int
				if octets, remote, err = server.ReadFromUDP(buf); err != nil {
					if !strings.Contains(err.Error(), "timeout") {
						astiLogger.Errorf("netflowd err=%v", err)
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
					astiLogger.Errorf("netflowd err=%v", err)
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
				astiLogger.Errorf("logIPFIX err=%v", err)
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
	if mapConf.SnmpMode != "" {
		tl.Params = &gosnmp.GoSNMP{}
		tl.Params.Version = gosnmp.Version3
		tl.Params.SecurityModel = gosnmp.UserSecurityModel
		if mapConf.SnmpMode == "v3auth" {
			tl.Params.MsgFlags = gosnmp.AuthNoPriv
			tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 mapConf.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: mapConf.Password,
			}
		} else {
			tl.Params.MsgFlags = gosnmp.AuthPriv
			tl.Params.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 mapConf.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: mapConf.Password,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        mapConf.Password,
			}
		}
	}
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
			astiLogger.Debug(err)
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
		astiLogger.Debug("Trap Listen End")
	}()
	for {
		select {
		case <-stopCh:
			{
				astiLogger.Debug("Trap Listen Done")
				return
			}
		}
	}
}
