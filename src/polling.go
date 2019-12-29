package main

/*
polling.go :ポーリング処理を行う
ポーリングの種類は
(1)能動的なポーリング
 ping
 snmp - sysUptime,ifOperStatus,
 http
 https
 tls
 dns
（２）受動的なポーリング
 syslog
 snmp trap
 netflow
 ipfix

*/

import (
	"context"
	"time"
	"strings"
	"fmt"
	"sort"
	"runtime"
	"net"

	astilog "github.com/asticode/go-astilog"

)

var (
	pollingStateChangeCh = make(chan *pollingEnt,10)
	doPollingCh = make(chan bool,10)
)

func pollingBackend(ctx context.Context) {
	go pingBackend(ctx)
	time.Sleep(time.Millisecond*100)
	var nextPoll int64
	for {
		select {
		case <-ctx.Done():
			return
		case  <- doPollingCh:
			{
				now := time.Now().UnixNano()
				if nextPoll > now {
					continue
				}
				nextPoll = now + (1000 * 1000 * 1000) * 2
				list := []*pollingEnt{}
				pollings.Range(func(_,v interface{}) bool {
					p := v.(*pollingEnt)
					if p.LastTime + (int64(p.PollInt) * 1000 * 1000 * 1000) < now {
						list = append(list,p)
					}
					return true
				})
				if len(list) < 1 {
					continue
				}
				astilog.Infof("doPolling %d NumGoroutine %d",len(list),runtime.NumGoroutine())
				sort.Slice(list,func (i,j int)bool {
					return list[i].LastTime < list[j].LastTime 
				})
				for i:=0; i < len(list);i++ {
					list[i].LastTime = time.Now().UnixNano()
					go doPolling(list[i])
					time.Sleep(time.Millisecond * 1)
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
			Event: fmt.Sprintf("ポーリング状態変化:%s(%s):%f:%s",p.Name,p.Type,p.LastVal,p.LastResult),
		})
	}
}

func doPolling(p *pollingEnt){
	oldState := p.State
	switch p.Type {
	case "ping":
		doPollingPing(p)
	case "snmp":
		doPollingSnmp(p)
	case "tcp":
		doPollingTCP(p)
	case "http","https":
		doPollingHTTP(p)
	case "tls":
		doPollingTLS(p)
	case "dns":
		doPollingDNS(p)
	case "syslog","trap","netflow","ipfix":
		doPollingLog(p)
		updatePolling(p)
	case "syslogpri":
		if !doPollingSyslogPri(p) {
			return
		}
	}
	if p.LogMode == 1 || p.LogMode == 3 || (p.LogMode == 2 && oldState != p.State) {
		if err := addPollingLog(p);err != nil {
			astilog.Errorf("addPollingLog err=%v",err)
		}
	}
}

func doPollingPing(p *pollingEnt){
	n,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	r := doPing(n.IP,p.Timeout,p.Retry,64)
	p.LastVal = float64(r.Time)
	if r.Stat == pingOK {
		p.LastResult = ""
		setPollingState(p,"normal")
	}	else {
		p.LastResult = fmt.Sprintf("%v",r.Error)
		setPollingState(p,p.Level)
	}
	updatePolling(p)
}

func doPollingDNS(p *pollingEnt){
	_,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	ok = false
	var rTime int64
	var ip string
	for i:=0  ; !ok && i <= p.Retry;i++{
		startTime := time.Now().UnixNano()
		addr, err := net.ResolveIPAddr("ip", p.Polling)
		endTime := time.Now().UnixNano()
		if err != nil {
			astilog.Debugf("doPollingDNS err=%v",err)
			p.LastResult = fmt.Sprintf("ERR:%v",err)
			continue
		}
		rTime = endTime - startTime
		ok = true
		ip = addr.String()
	}
	if ok && p.LastResult != ""  && !strings.HasPrefix(p.LastResult,"ERR") && ip != p.LastResult {
		ok = false
	}
	p.LastVal = float64(rTime)
	if ok {
		p.LastResult = ip
		setPollingState(p,"normal")
	}	else {
		setPollingState(p,p.Level)
	}
	updatePolling(p)
}

