package main

/* discover.go: 自動発見の処理
	自動発見は、PINGを実行して、応答があるノードに関してSNMPの応答があるか確認する
*/

import (
	"fmt"
	"time"
	"net"
	"strings"
	"github.com/signalsciences/ipv4"
	"github.com/soniah/gosnmp"
	astilog "github.com/asticode/go-astilog"
)

type discoverStatEnt struct {
	Running bool
	Stop  bool
	Total uint32
	Sent  uint32
	Found uint32 
	Snmp  uint32
	Progress uint32
	StartTime int64
	EndTime  int64
	X int
	Y int
}

type discoverInfoEnt struct {
	IP string
	HostName string
	SysName string
	SysObjectID string
	IfIndexList []string
}

var discoverStat discoverStatEnt

func stopDiscover() {
	for discoverStat.Running {
		discoverStat.Stop = true
		time.Sleep(time.Millisecond*100)
	}
}

func startDiscover() error {
	if discoverStat.Running {
		return fmt.Errorf("Discover already runnning")
	}
	sip, err := ipv4.FromDots(discoverConf.StartIP)
	if err != nil {
		return fmt.Errorf("Discover StartIP err=%v",err)
	}
	eip, err := ipv4.FromDots(discoverConf.EndIP)
	if err != nil {
		return fmt.Errorf("Discover EndIP err=%v",err)
	}
	if sip > eip {
		return fmt.Errorf("Discover StartIP > EndIP")
	}
	astilog.Debug("Start doDiscover")
	addEventLog(eventLogEnt{
		Type: "system",
		Level:"info",
		Event: fmt.Sprintf("自動発見開始 %s - %s",discoverConf.StartIP,discoverConf.EndIP),
	})
	discoverStat.Stop = false
	discoverStat.Total = eip - sip
	discoverStat.Sent  = 0
	discoverStat.Found = 0 
	discoverStat.Snmp  = 0
	discoverStat.Running = true
	discoverStat.StartTime = time.Now().UnixNano()
	discoverStat.EndTime = 0
	discoverStat.X = discoverConf.X
	discoverStat.Y = discoverConf.Y
	sem := make(chan bool, 10)
	go func() {
		for ; sip < eip && !discoverStat.Stop ;sip++ {
			sem <- true
			discoverStat.Sent++
			discoverStat.Progress = (100 * discoverStat.Sent)/discoverStat.Total
			go func(ip uint32) {
				defer func() {
					<-sem
				}()
				ipstr := ipv4.ToDots(ip)
				if findNodeFromIP(ipstr) != nil {
					return
				}
				r := doPing(ipstr,1,0,64)
				if r.Stat == pingOK {
					discoverStat.Found++
					dent := discoverInfoEnt{
						IP: ipstr,
						IfIndexList: []string{},
					}
					if names,err := net.LookupAddr(ipstr); err !=nil && len(names) > 0 {
						dent.HostName = names[0]
					}
					discoverGetSnmpInfo(ipstr,&dent)
					if dent.SysName != "" {
						discoverStat.Snmp++
					}
					go addFoundNode(dent)
					discoverStat.X += 96
					if discoverStat.X > 1024 {
						discoverStat.X = 32
						discoverStat.Y += 64
					}
				}
			}(sip)
		}
		for len(sem) > 0 {
			time.Sleep(time.Millisecond * 10)
		}
		discoverStat.Running = false
		discoverStat.EndTime = time.Now().UnixNano()
		addEventLog(eventLogEnt{
			Type: "system",
			Level:"info",
			Event: fmt.Sprintf("自動発見終了 %s - %s",discoverConf.StartIP,discoverConf.EndIP),
		})
		doPollingCh <- true
	}()
	return nil
}

func discoverGetSnmpInfo(t string,dent *discoverInfoEnt) {
	agent := &gosnmp.GoSNMP{
		Target:             t,
		Port:               161,
		Transport:          "udp",
		Community:          discoverConf.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(2) * time.Second,
		Retries:            1,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	err := agent.Connect()
	if err != nil {
		astilog.Errorf("discoverGetSnmpInfo err=%v",err)
		return
	}
	defer agent.Conn.Close()
	oids := []string{mib.NameToOID("sysName"), mib.NameToOID("sysObjectID")}
	result, err := agent.GetNext(oids)
	if err != nil {
		astilog.Errorf("discoverGetSnmpInfo err=%v",err)
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
		a := strings.Split(mib.OIDToName(variable.Name),".")
		if len(a) == 2 && 
			a[0] == "ifType"  && 
			gosnmp.ToBigInt(variable.Value).Int64() == 6 {
			dent.IfIndexList = append(dent.IfIndexList,a[1])
		}
		return nil
	})
	return
}

func addFoundNode(dent discoverInfoEnt) {
	n := nodeEnt{
		Name: dent.HostName,
		IP: dent.IP,
		Icon: "desktop",
		X: discoverStat.X,
		Y: discoverStat.Y,
		Descr: "自動登録:" + time.Now().Format(time.RFC3339),
	}
	if n.Name == "" {
		if dent.SysName != "" {
			n.Name = dent.SysName
		} else {
			n.Name = dent.IP
		}
	}
	if dent.SysObjectID != ""{
		n.Community = discoverConf.Community
		n.Icon = "hdd"
	}
	if err := addNode(&n); err != nil{
		astilog.Error(err)
		return
	}
	addEventLog(eventLogEnt{
		Type:"discover",
		Level:"info",
		NodeID: n.ID,
		NodeName: n.Name,
		Event: "自動発見により追加",
	})
	p := &pollingEnt{
		NodeID: n.ID,
		Name: "PING監視",
		Type: "ping",
		Level: "low",
		State: "unkown",
		PollInt: mapConf.PollInt,
		Timeout: mapConf.Timeout,
		Retry: mapConf.Retry,
	}
	if err := addPolling(p); err != nil{
		astilog.Error(err)
		return
	}
	if dent.SysObjectID == "" {
		return
	}
	p = &pollingEnt{
		NodeID: n.ID,
		Name: "sysUptime監視",
		Type: "snmp",
		Polling: "sysUpTime",
		Level: "low",
		State: "unkown",
		PollInt: mapConf.PollInt,
		Timeout: mapConf.Timeout,
		Retry: mapConf.Retry,
	}
	if err := addPolling(p); err != nil{
		astilog.Error(err)
		return
	}
	for _,i := range dent.IfIndexList {
		p = &pollingEnt{
			NodeID: n.ID,
			Type: "snmp",
			Name: "IF " + i + "監視",
			Polling: "ifOperStatus." + i,
			Level: "low",
			State: "unkown",
			PollInt: mapConf.PollInt,
			Timeout: mapConf.Timeout,
			Retry: mapConf.Retry,
		}
		if err := addPolling(p); err != nil{
			astilog.Error(err)
			return
		}
	}
}