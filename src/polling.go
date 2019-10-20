package main

/*
polling.go :ポーリング処理を行う
ポーリングの種類は
(1)能動的なポーリング
 ping
 snmp - sysUptime,ifOperStatus,
 wget
（２）受動的なポーリング
 syslog
 trap
 netflow

*/

import (
	"context"
	"encoding/json"
	"time"
	"strings"
	"fmt"
	"strconv"
	"sort"

	gosnmp "github.com/soniah/gosnmp"

	astilog "github.com/asticode/go-astilog"
	ping "github.com/sparrc/go-ping"

)

var (
	pollingStateChangeCh = make(chan *pollingEnt,10)
)

func pollingBackend(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			{
				return
			}
		case <-time.Tick(time.Second * 10):
			{
				list := []*pollingEnt{}
				for _,p := range pollings {
					if p.LastTime + (int64(p.PollInt) * 1000 * 1000 * 1000) < time.Now().UnixNano() {
						list = append(list,p)
					}
				}
				sort.Slice(list,func (i,j int)bool {
					return list[i].LastTime < list[j].LastTime 
				})
				for i:=0; i < 100 && i < len(list);i++ {
					list[i].LastTime = time.Now().UnixNano()
					go doPolling(list[i])
				}
			}
		}
	}
}

func setPollingState(p *pollingEnt,newState string){
	sendEvent := false
	if newState == "normal" {
		if p.State != "normal" && p.State != "repair" {
			if p.State == "unkown"{
				p.State = "normal"
			} else {
				p.State = "repair"
			}
			sendEvent = true
		}
	} else if newState == "unkown" {
		if p.State != "unkown" {
			p.State = "unkown"
			sendEvent = true
		}
	}	else {
		if p.State != p.Level {
			p.State = p.Level
			sendEvent = true
		}
	}
	if sendEvent {
		nodeName := "Unknown"
		if n,ok := nodes[p.NodeID]; ok {
			nodeName = n.Name
		}
		pollingStateChangeCh <- p
		addEventLog(eventLogEnt{
			Type:"polling",
			Level: p.State,
			NodeID: p.NodeID,
			NodeName: nodeName,
			Event: "ポーリング状態変化:" + p.Name,
		})
	}
}

func doPolling(p *pollingEnt){
	switch p.Type {
	case "ping":
		doPollingPing(p)
	case "snmp":
		doPollingSnmp(p)
	}
}

func doPollingPing(p *pollingEnt){
	n,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	pinger, err := ping.NewPinger(n.IP)
	if err != nil {
		astilog.Errorf("NewPinger err=%v",err)
		return
	}
	pinger.Count = 1
	pinger.Timeout = time.Second * time.Duration(p.Timeout)
	pinger.Run()
	s := pinger.Statistics()
	if s.PacketsRecv > 0 {
		setPollingState(p,"normal")
	}	else {
		setPollingState(p,p.Level)
	}
	js,err := json.Marshal(&s)
	if err != nil {
		astilog.Errorf("ping Marshal err=%",err)
		return
	}
	p.LastVal = s.AvgRtt.Nanoseconds()
	p.LastResult = string(js)
	updatePolling(p)
	if strings.Contains(p.Polling,"log"){
		addPollingLog(p)
	}
}

func doPollingSnmp(p *pollingEnt){
	n,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(p.Timeout) * time.Second,
		Retries:            p.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	err := agent.Connect()
	if err != nil {
		astilog.Errorf("SNMP agent.Connect err=%v",err)
		return
	}
	defer agent.Conn.Close()
	ps,mode,log := parseSnmpPolling(p.Polling)
	if ps == "" {
		astilog.Errorf("Empty SNMP Polling %s",p.Name)
		return
	}
	if ps == "sysUpTime" {
		doPollingSnmpSysUpTime(p,agent)
	} else if strings.HasPrefix(ps,"ifOperStatus.") {
		doPollingSnmpIF(p,ps,agent)
	} else {
		doPollingSnmpOther(p,ps,mode,agent)
	}
	updatePolling(p)
	if log {
		addPollingLog(p)
	}
}

func parseSnmpPolling(s string) (string,string,bool) {
	a :=  strings.Split(s,"|")
	if len(a) < 1 {
		return "","",false
	}
	ps := strings.TrimSpace(a[0])
	if len(a) < 2 {
		return ps,"",false
	}
	mode := strings.TrimSpace(a[1])
	if len(a) < 3 {
		if mode == "log" {
			return ps,"",true
		}
		return ps,mode,false
	}
	if strings.TrimSpace(a[1]) == "log"{
		return ps,mode, true
	}
	return ps,mode, false
}

func doPollingSnmpSysUpTime(p *pollingEnt,agent *gosnmp.GoSNMP){
	oids := []string{mib.NameToOID("sysUpTime.0")}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingState(p,"unkown")
		return
	}
	var uptime int64
	for _, variable := range result.Variables {
		if variable.Name == mib.NameToOID("sysUpTime.0") {
			uptime = gosnmp.ToBigInt(variable.Value).Int64()
			break
		}
	}
	if uptime == 0 {
		setPollingState(p,"unkown")
		return
	}
	p.LastVal = uptime
	if p.LastResult == "" {
		p.LastResult = fmt.Sprintf("%d",uptime)
		return
	}
	if lastUptime,err := strconv.ParseInt(p.LastResult,10,64);err != nil {
		p.LastResult = fmt.Sprintf("%d",uptime)
		setPollingState(p,"unkown")
	} else {
		p.LastResult = fmt.Sprintf("%d",uptime)
		if lastUptime < uptime {
			setPollingState(p,"normal")
			return
		}
		setPollingState(p,p.Level)
	}
}

func doPollingSnmpIF(p *pollingEnt,ps string,agent *gosnmp.GoSNMP) {
	a := strings.Split(ps,".")
	if len(a) < 2 {
		setPollingState(p,"unkown")
		return
	}
	oids := []string{mib.NameToOID("ifOperStatus."+a[1]),mib.NameToOID("ifAdminState."+a[1])}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingState(p,"unkown")
		return
	}
	var oper int64
	var admin int64
	for _, variable := range result.Variables {
		if strings.HasPrefix(mib.OIDToName(variable.Name),"ifOperStatus") {
			oper = gosnmp.ToBigInt(variable.Value).Int64()
		} else if strings.HasPrefix(mib.OIDToName(variable.Name),"ifAdminStatus") {
			admin = gosnmp.ToBigInt(variable.Value).Int64()
		}
	}
	p.LastVal = oper
	p.LastResult = fmt.Sprintf("{oper:%d, admin:%d}",oper,admin)
	if oper == 1 {
		setPollingState(p,"normal")
		return
	} else if admin == 2 {
		setPollingState(p,"normal")
		return
	} else if oper == 2 && admin == 1 {
		setPollingState(p,p.Level)
		return
	}
	astilog.Debugf("SNMP IF Polling admin=%d oper=%d",admin,oper)
	setPollingState(p,"unkown")
	return
}

func doPollingSnmpOther(p *pollingEnt,ps,mode string,agent *gosnmp.GoSNMP) {
	a := strings.Split(ps," ")
	if len(a) < 3 {
		setPollingState(p,"unkown")
		return
	}
	m := strings.TrimSpace(a[0])
	op := strings.TrimSpace(a[1])
	cv := strings.TrimSpace(a[2])
	oids := []string{mib.NameToOID(m)}
	if mode == "ps" {
		oids = append(oids,mib.NameToOID("sysUpTime.0"))
	}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingState(p,"unkown")
		return
	}
	var iv int64
	var sut int64
	var sv string
	hitIv := false
	hitSv := false
	for _, variable := range result.Variables {
		if variable.Name == mib.NameToOID(ps) {
			if variable.Type == gosnmp.OctetString {
				sv = string(variable.Value.([]byte))
				hitSv = true
			} else if variable.Type == gosnmp.ObjectIdentifier {
				sv = mib.OIDToName(variable.Value.(string))
				hitSv = true
			} else {
				iv = gosnmp.ToBigInt(variable.Value).Int64()
				hitIv = true
			}
		} else if variable.Name == mib.NameToOID("sysUpTime.0"){
			sut = gosnmp.ToBigInt(variable.Value).Int64()
		}
	}
	if !hitIv && !hitSv {
		setPollingState(p,"unkown")
		return
	}
	if hitIv {
		sv = fmt.Sprintf("%d,%d",iv,sut)
	}
	if mode == "ps" || mode == "delta" {
		if p.LastResult == "" {
			p.LastResult =  sv
			return
		}
	}
	r := false
	if hitSv {
		switch op {
		case "=","==":
			r = sv == cv 
		case "~=":
			r = strings.Contains(sv,cv)
		case "<":
			r = strings.Compare(sv,cv) < 0
		case ">":
			r = strings.Compare(sv,cv) > 0
		default:
			setPollingState(p,"unkown")
			return
		}
		p.LastResult = sv 
	} else {
		civ,err :=  strconv.ParseInt(cv,10,64)
		if err != nil {
			p.LastResult = fmt.Sprintf("%d,%d",iv,sut)
			setPollingState(p,"unkown")
			return
		}
		var liv int64
		var lsut int64
		n,err :=  fmt.Sscanf(p.LastResult,"%d,%d",&liv,lsut)
		if err != nil || n != 2 {
			p.LastResult = fmt.Sprintf("%d,%d",iv,sut)
			setPollingState(p,"unkown")
			return
		}
		if mode == "ps" {
			dsut := sut -  lsut
			if dsut <= 0 {
				p.LastResult = fmt.Sprintf("%d,%d",iv,sut)
				setPollingState(p,"unkown")
				return
			}
			iv = (100*(iv-liv))/dsut
		} else if mode == "delta" {
			iv -= liv
		}
		switch op {
		case "=","==":
			r = iv == civ 
		case "!=":
			r = iv != civ 
		case "<":
			r =  iv < civ
		case ">":
			r = iv > civ
		case "<=":
			r =  iv < civ
		case ">=":
			r = iv >= civ
		default:
			setPollingState(p,"unkown")
			return
		}
		p.LastVal = iv
	}
	if r {
		setPollingState(p,"normal")
		return
	}
	setPollingState(p,p.Level)
	return
}


func addPollingLog(p *pollingEnt) {
	s, err := json.Marshal(p)
	if err != nil {
		astilog.Errorf("polling Marshal err=%v",err)
		return
	}
	logCh <- &logEnt{
		Time: time.Now().UnixNano(),
		Type: "pollingLogs",
		Log:  string(s),
	}
}