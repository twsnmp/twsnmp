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
	if js,err := json.Marshal(&s); err != nil {
		p.LastResult = string(js)
		astilog.Debugf("ping=%s",p.LastResult)
	}
	updatePolling(p)
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
	if p.Polling == "sysUpTime" {
		doSnmpPollingSysUpTime(p,agent)
	} else if strings.HasPrefix(p.Polling,"ifOperStatus.") {
		doSnmpPollingIF(p,agent)
	}
	updatePolling(p)
}

func doSnmpPollingSysUpTime(p *pollingEnt,agent *gosnmp.GoSNMP){
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

func doSnmpPollingIF(p *pollingEnt,agent *gosnmp.GoSNMP) {
	a := strings.Split(p.Polling,".")
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
	if oper == 1 {
		setPollingState(p,"normal")
		return
	} else if admin == 2 {
		setPollingState(p,"normal")
		return
	} else if oper == 2 {
		setPollingState(p,p.Level)
		return
	}
	setPollingState(p,"unkown")
	return
}
