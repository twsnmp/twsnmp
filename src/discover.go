package main

/* discover.go: 自動発見の処理
自動発見は、PINGを実行して、応答があるノードに関してSNMPの応答があるか確認する
*/

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/signalsciences/ipv4"
	"github.com/soniah/gosnmp"
)

type discoverStatEnt struct {
	Running   bool
	Stop      bool
	Total     uint32
	Sent      uint32
	Found     uint32
	Snmp      uint32
	Progress  uint32
	StartTime int64
	EndTime   int64
	X         int
	Y         int
}

type discoverInfoEnt struct {
	IP          string
	HostName    string
	SysName     string
	SysObjectID string
	IfIndexList []string
	X           int
	Y           int
}

var discoverStat discoverStatEnt

func stopDiscover() {
	for discoverStat.Running {
		discoverStat.Stop = true
		time.Sleep(time.Millisecond * 100)
	}
}

// GRID : 自動発見時にノードを配置する間隔
const GRID = 90

func startDiscover() error {
	if discoverStat.Running {
		return fmt.Errorf("Discover already runnning")
	}
	sip, err := ipv4.FromDots(discoverConf.StartIP)
	if err != nil {
		return fmt.Errorf("Discover StartIP err=%v", err)
	}
	eip, err := ipv4.FromDots(discoverConf.EndIP)
	if err != nil {
		return fmt.Errorf("Discover EndIP err=%v", err)
	}
	if sip > eip {
		return fmt.Errorf("Discover StartIP > EndIP")
	}
	addEventLog(eventLogEnt{
		Type:  "system",
		Level: "info",
		Event: fmt.Sprintf("自動発見開始 %s - %s", discoverConf.StartIP, discoverConf.EndIP),
	})
	discoverStat.Stop = false
	discoverStat.Total = eip - sip + 1
	discoverStat.Sent = 0
	discoverStat.Found = 0
	discoverStat.Snmp = 0
	discoverStat.Running = true
	discoverStat.StartTime = time.Now().UnixNano()
	discoverStat.EndTime = 0
	discoverStat.X = (1 + discoverConf.X/GRID) * GRID
	discoverStat.Y = (1 + discoverConf.Y/GRID) * GRID
	var mu sync.Mutex
	sem := make(chan bool, 20)
	go func() {
		for ; sip <= eip && !discoverStat.Stop; sip++ {
			sem <- true
			discoverStat.Sent++
			discoverStat.Progress = (100 * discoverStat.Sent) / discoverStat.Total
			go func(ip uint32) {
				defer func() {
					<-sem
				}()
				ipstr := ipv4.ToDots(ip)
				if findNodeFromIP(ipstr) != nil {
					return
				}
				r := doPing(ipstr, discoverConf.Timeout, discoverConf.Retry, 64)
				if r.Stat == pingOK {
					dent := discoverInfoEnt{
						IP:          ipstr,
						IfIndexList: []string{},
					}
					if names, err := net.LookupAddr(ipstr); err == nil && len(names) > 0 {
						dent.HostName = names[0]
					}
					discoverGetSnmpInfo(ipstr, &dent)
					mu.Lock()
					dent.X = discoverStat.X
					dent.Y = discoverStat.Y
					discoverStat.Found++
					discoverStat.X += GRID
					if discoverStat.X > GRID*10 {
						discoverStat.X = GRID
						discoverStat.Y += GRID
					}
					if dent.SysName != "" {
						discoverStat.Snmp++
					}
					addFoundNode(dent)
					mu.Unlock()
				}
			}(sip)
		}
		for len(sem) > 0 {
			time.Sleep(time.Millisecond * 10)
		}
		discoverStat.Running = false
		discoverStat.EndTime = time.Now().UnixNano()
		addEventLog(eventLogEnt{
			Type:  "system",
			Level: "info",
			Event: fmt.Sprintf("自動発見終了 %s - %s", discoverConf.StartIP, discoverConf.EndIP),
		})
		doPollingCh <- true
	}()
	return nil
}

func discoverGetSnmpInfo(t string, dent *discoverInfoEnt) {
	agent := &gosnmp.GoSNMP{
		Target:             t,
		Port:               161,
		Transport:          "udp",
		Community:          mapConf.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(2) * time.Second,
		Retries:            1,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	if discoverConf.SnmpMode != "" {
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		if mapConf.SnmpMode == "v3auth" {
			agent.MsgFlags = gosnmp.AuthNoPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 mapConf.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: mapConf.Password,
			}
		} else {
			agent.MsgFlags = gosnmp.AuthPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 mapConf.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: mapConf.Password,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        mapConf.Password,
			}
		}
	}
	err := agent.Connect()
	if err != nil {
		astiLogger.Errorf("discoverGetSnmpInfo err=%v", err)
		return
	}
	defer agent.Conn.Close()
	oids := []string{mib.NameToOID("sysName"), mib.NameToOID("sysObjectID")}
	result, err := agent.GetNext(oids)
	if err != nil {
		astiLogger.Errorf("discoverGetSnmpInfo err=%v", err)
		return
	}
	for _, variable := range result.Variables {
		if mib.OIDToName(variable.Name) == "sysName.0" {
			dent.SysName = string(variable.Value.([]byte))
		} else if mib.OIDToName(variable.Name) == "sysObjectID.0" {
			dent.SysObjectID = variable.Value.(string)
		}
	}
	err = agent.Walk(mib.NameToOID("ifType"), func(variable gosnmp.SnmpPDU) error {
		a := strings.Split(mib.OIDToName(variable.Name), ".")
		if len(a) == 2 &&
			a[0] == "ifType" &&
			gosnmp.ToBigInt(variable.Value).Int64() == 6 {
			dent.IfIndexList = append(dent.IfIndexList, a[1])
		}
		return nil
	})
	return
}

func addFoundNode(dent discoverInfoEnt) {
	n := nodeEnt{
		Name:  dent.HostName,
		IP:    dent.IP,
		Icon:  "desktop",
		X:     dent.X,
		Y:     dent.Y,
		Descr: "自動登録:" + time.Now().Format(time.RFC3339),
	}
	if n.Name == "" {
		if dent.SysName != "" {
			n.Name = dent.SysName
		} else {
			n.Name = dent.IP
		}
	}
	if dent.SysObjectID != "" {
		n.SnmpMode = mapConf.SnmpMode
		n.User = mapConf.User
		n.Password = mapConf.Password
		n.Community = mapConf.Community
		n.Icon = "hdd"
	}
	if err := addNode(&n); err != nil {
		astiLogger.Error(err)
		return
	}
	addEventLog(eventLogEnt{
		Type:     "discover",
		Level:    "info",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    "自動発見により追加",
	})
	p := &pollingEnt{
		NodeID:  n.ID,
		Name:    "PING監視",
		Type:    "ping",
		Level:   "low",
		State:   "unknown",
		PollInt: mapConf.PollInt,
		Timeout: mapConf.Timeout,
		Retry:   mapConf.Retry,
	}
	if err := addPolling(p); err != nil {
		astiLogger.Error(err)
		return
	}
	if dent.SysObjectID == "" {
		return
	}
	p = &pollingEnt{
		NodeID:  n.ID,
		Name:    "sysUptime監視",
		Type:    "snmp",
		Polling: "sysUpTime",
		Level:   "low",
		State:   "unknown",
		PollInt: mapConf.PollInt,
		Timeout: mapConf.Timeout,
		Retry:   mapConf.Retry,
	}
	if err := addPolling(p); err != nil {
		astiLogger.Error(err)
		return
	}
	for _, i := range dent.IfIndexList {
		p = &pollingEnt{
			NodeID:  n.ID,
			Type:    "snmp",
			Name:    "IF " + i + "監視",
			Polling: "ifOperStatus." + i,
			Level:   "low",
			State:   "unknown",
			PollInt: mapConf.PollInt,
			Timeout: mapConf.Timeout,
			Retry:   mapConf.Retry,
		}
		if err := addPolling(p); err != nil {
			astiLogger.Error(err)
			return
		}
	}
}
