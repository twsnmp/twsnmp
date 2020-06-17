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
	"fmt"
	"net"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/beevik/ntp"
)

var (
	pollingStateChangeCh = make(chan *pollingEnt, 10)
	doPollingCh          = make(chan bool, 10)
)

const (
	logModeNone = iota
	logModeAlways
	logModeOnChange
	logModeAI
)

func pollingBackend(ctx context.Context) {
	loadGrokMap()
	go pingBackend(ctx)
	time.Sleep(time.Millisecond * 100)
	var nextPoll int64
	for {
		select {
		case <-ctx.Done():
			return
		case <-doPollingCh:
			{
				now := time.Now().UnixNano()
				if nextPoll > now {
					continue
				}
				nextPoll = now + (1000*1000*1000)*2
				list := []*pollingEnt{}
				pollings.Range(func(_, v interface{}) bool {
					p := v.(*pollingEnt)
					if p.NextTime < now {
						list = append(list, p)
					}
					return true
				})
				if len(list) < 1 {
					continue
				}
				astiLogger.Infof("doPolling %d NumGoroutine %d", len(list), runtime.NumGoroutine())
				sort.Slice(list, func(i, j int) bool {
					return list[i].NextTime < list[j].NextTime
				})
				for i := 0; i < len(list); i++ {
					list[i].NextTime = time.Now().UnixNano() + (int64(list[i].PollInt) * 1000 * 1000 * 1000)
					go doPolling(list[i])
					time.Sleep(time.Millisecond * 1)
				}
			}
		}
	}
}

func setPollingState(p *pollingEnt, newState string) {
	sendEvent := false
	if newState == "normal" {
		if p.State != "normal" && p.State != "repair" {
			if p.State == "unkown" {
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
	} else {
		if p.State != p.Level {
			p.State = p.Level
			sendEvent = true
		}
	}
	if sendEvent {
		nodeName := "Unknown"
		if n, ok := nodes[p.NodeID]; ok {
			nodeName = n.Name
		}
		pollingStateChangeCh <- p
		addEventLog(eventLogEnt{
			Type:     "polling",
			Level:    p.State,
			NodeID:   p.NodeID,
			NodeName: nodeName,
			Event:    fmt.Sprintf("ポーリング状態変化:%s(%s):%f:%s", p.Name, p.Type, p.LastVal, p.LastResult),
		})
	}
}

func doPolling(p *pollingEnt) {
	oldState := p.State
	switch p.Type {
	case "ping":
		doPollingPing(p)
	case "snmp":
		doPollingSnmp(p)
	case "tcp":
		doPollingTCP(p)
	case "http", "https":
		doPollingHTTP(p)
	case "tls":
		doPollingTLS(p)
	case "dns":
		doPollingDNS(p)
	case "ntp":
		doPollingNTP(p)
	case "syslog", "trap", "netflow", "ipfix":
		doPollingLog(p)
		updatePolling(p)
	case "syslogpri":
		if !doPollingSyslogPri(p) {
			return
		}
	case "syslogdevice":
		doPollingSyslogDevice(p)
		updatePolling(p)
	case "sysloguser":
		doPollingSyslogUser(p)
		updatePolling(p)
	case "syslogflow":
		doPollingSyslogFlow(p)
		updatePolling(p)
	case "cmd":
		doPollingCmd(p)
		updatePolling(p)
	}
	if p.LogMode == logModeAlways || p.LogMode == logModeAI || (p.LogMode == logModeOnChange && oldState != p.State) {
		if err := addPollingLog(p); err != nil {
			astiLogger.Errorf("addPollingLog err=%v", err)
		}
	}
}

func doPollingPing(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	r := doPing(n.IP, p.Timeout, p.Retry, 64)
	p.LastVal = float64(r.Time)
	if r.Stat == pingOK {
		p.LastResult = ""
		setPollingState(p, "normal")
	} else {
		p.LastResult = fmt.Sprintf("%v", r.Error)
		setPollingState(p, p.Level)
	}
	updatePolling(p)
}

func doPollingDNS(p *pollingEnt) {
	_, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	ok = false
	var rTime int64
	var ip string
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		addr, err := net.ResolveIPAddr("ip", p.Polling)
		endTime := time.Now().UnixNano()
		if err != nil {
			astiLogger.Debugf("doPollingDNS err=%v", err)
			p.LastResult = fmt.Sprintf("ERR:%v", err)
			continue
		}
		rTime = endTime - startTime
		ok = true
		ip = addr.String()
	}
	if ok && p.LastResult != "" && !strings.HasPrefix(p.LastResult, "ERR") && ip != p.LastResult {
		ok = false
	}
	p.LastVal = float64(rTime)
	if ok {
		p.LastResult = ip
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	updatePolling(p)
}

func doPollingNTP(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	ok = false
	for i := 0; !ok && i <= p.Retry; i++ {
		options := ntp.QueryOptions{Timeout: time.Duration(p.Timeout) * time.Second}
		r, err := ntp.QueryWithOptions(n.IP, options)
		if err != nil {
			astiLogger.Debugf("doPollingNTP err=%v", err)
			p.LastResult = fmt.Sprintf("%v", err)
			continue
		}
		p.LastVal = float64(r.RTT.Nanoseconds())
		p.LastResult = fmt.Sprintf("Stratum=%d;ReferenceID=%d;ClockOffset=%d", r.Stratum, r.ReferenceID, r.ClockOffset.Nanoseconds())
		ok = true
	}
	if ok {
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	updatePolling(p)
}
