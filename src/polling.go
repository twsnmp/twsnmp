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
	"encoding/json"
	"fmt"
	"net"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/beevik/ntp"
	"github.com/robertkrimen/otto"
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
	for {
		select {
		case <-ctx.Done():
			return
		case <-doPollingCh:
			{
				now := time.Now().UnixNano()
				list := []*pollingEnt{}
				pollings.Range(func(_, v interface{}) bool {
					p := v.(*pollingEnt)
					if p.NextTime < (now + (10 * 1000 * 1000 * 1000)) {
						list = append(list, p)
					}
					return true
				})
				if len(list) < 1 {
					continue
				}
				astiLogger.Infof("doPolling=%d NumGoroutine=%d", len(list), runtime.NumGoroutine())
				sort.Slice(list, func(i, j int) bool {
					return list[i].NextTime < list[j].NextTime
				})
				for i := 0; i < len(list); i++ {
					startTime := list[i].NextTime
					if startTime < now {
						startTime = now
					}
					list[i].NextTime = startTime + (int64(list[i].PollInt) * 1000 * 1000 * 1000)
					go doPolling(list[i], startTime)
					time.Sleep(time.Millisecond * 2)
				}
			}
		}
	}
}

func setPollingState(p *pollingEnt, newState string) {
	sendEvent := false
	if newState == "normal" {
		if p.State != "normal" && p.State != "repair" {
			if p.State == "unknown" {
				p.State = "normal"
			} else {
				p.State = "repair"
			}
			sendEvent = true
		}
	} else if newState == "unknown" {
		if p.State != "unknown" {
			p.State = "unknown"
			sendEvent = true
		}
	} else {
		if p.State != p.Level {
			p.State = p.Level
			sendEvent = true
		}
	}
	if sendEvent {
		nodeName := "unknown"
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

func doPolling(p *pollingEnt, startTime int64) {
	for startTime > time.Now().UnixNano() {
		time.Sleep(time.Millisecond * 100)
	}
	oldState := p.State
	switch p.Type {
	case "ping":
		doPollingPing(p)
		updatePolling(p)
	case "snmp":
		doPollingSnmp(p)
		updatePolling(p)
	case "tcp":
		doPollingTCP(p)
		updatePolling(p)
	case "http", "https":
		doPollingHTTP(p)
		updatePolling(p)
	case "tls":
		doPollingTLS(p)
		updatePolling(p)
	case "dns":
		doPollingDNS(p)
		updatePolling(p)
	case "ntp":
		doPollingNTP(p)
		updatePolling(p)
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
	case "ssh":
		doPollingSSH(p)
		updatePolling(p)
	case "vmware":
		doPollingVmWare(p)
		updatePolling(p)
	case "twsnmp":
		doPollingTWSNMP(p)
		updatePolling(p)
	}
	if p.LogMode == logModeAlways || p.LogMode == logModeAI || (p.LogMode == logModeOnChange && oldState != p.State) {
		if err := addPollingLog(p); err != nil {
			astiLogger.Errorf("addPollingLog err=%v %#v", err, p)
		}
	}
	if influxdbConf.PollingLog != "" {
		if influxdbConf.PollingLog == "all" || p.LogMode != logModeNone {
			sendPollingLogToInfluxdb(p)
		}
	}
}

func doPollingPing(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		setPollingError("ping", p, fmt.Errorf("Node not found"))
		return
	}
	lr := make(map[string]string)
	r := doPing(n.IP, p.Timeout, p.Retry, 64)
	p.LastVal = float64(r.Time)
	if r.Stat == pingOK {
		lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
		setPollingState(p, "normal")
	} else {
		lr["error"] = fmt.Sprintf("%v", r.Error)
		setPollingState(p, p.Level)
	}
	p.LastResult = makeLastResult(lr)
}

func doPollingDNS(p *pollingEnt) {
	_, ok := nodes[p.NodeID]
	if !ok {
		setPollingError("dns", p, fmt.Errorf("Node not dond"))
		return
	}
	cmds := splitCmd(p.Polling)
	mode := "ipaddr"
	target := p.Polling
	script := ""
	if len(cmds) == 3 {
		mode = cmds[0]
		target = cmds[1]
		script = cmds[2]
	}
	ok = false
	var rTime int64
	var out []string
	var err error
	lr := make(map[string]string)
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		if out, err = doLookup(mode, target); err != nil || len(out) < 1 {
			lr["error"] = fmt.Sprintf("%v", err)
			astiLogger.Errorf("doLookup err=%v %v", err, cmds)
			continue
		}
		endTime := time.Now().UnixNano()
		rTime = endTime - startTime
		ok = true
		delete(lr, "error")
	}
	oldlr := make(map[string]string)
	json.Unmarshal([]byte(p.LastResult), &oldlr)
	if !ok {
		for k, v := range oldlr {
			if k != "error" {
				lr[k] = v
			}
		}
		p.LastResult = makeLastResult(lr)
		p.LastVal = 0.0
		setPollingState(p, p.Level)
		return
	}
	p.LastVal = float64(rTime)
	vm := otto.New()
	vm.Set("rtt", p.LastVal)
	vm.Set("count", len(out))
	lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
	lr["count"] = fmt.Sprintf("%d", len(out))
	switch mode {
	case "ipaddr":
		lr["ip"] = out[0]
		p.LastResult = makeLastResult(lr)
		if oldlr["ip"] != "" && oldlr["ip"] != lr["ip"] {
			setPollingState(p, p.Level)
			return
		}
		setPollingState(p, "normal")
		return
	case "addr":
		vm.Set("addr", out)
		lr["addr"] = strings.Join(out, ",")
	case "host":
		vm.Set("host", out)
		lr["host"] = strings.Join(out, ",")
	case "mx":
		vm.Set("mx", out)
		lr["mx"] = strings.Join(out, ",")
	case "ns":
		vm.Set("ns", out)
		lr["ns"] = strings.Join(out, ",")
	case "txt":
		vm.Set("txt", out)
		lr["txt"] = strings.Join(out, ",")
	case "cname":
		vm.Set("cname", out[0])
		lr["cname"] = out[0]
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("dns", p, fmt.Errorf("%v", err))
		return
	}
	p.LastResult = makeLastResult(lr)
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(p, "normal")
		return
	}
	setPollingState(p, p.Level)
	return
}

func doLookup(mode, target string) ([]string, error) {
	ret := []string{}
	switch mode {
	case "ipaddr":
		if addr, err := net.ResolveIPAddr("ip", target); err == nil {
			return []string{addr.String()}, nil
		} else {
			return ret, err
		}
	case "addr":
		return net.LookupAddr(target)
	case "host":
		return net.LookupHost(target)
	case "mx":
		if mxs, err := net.LookupMX(target); err == nil {
			for _, mx := range mxs {
				ret = append(ret, mx.Host)
			}
			return ret, nil
		} else {
			return ret, err
		}
	case "ns":
		if nss, err := net.LookupNS(target); err == nil {
			for _, ns := range nss {
				ret = append(ret, ns.Host)
			}
			return ret, nil
		} else {
			return ret, err
		}
	case "cname":
		if cname, err := net.LookupCNAME(target); err == nil {
			return []string{cname}, nil
		} else {
			return ret, err
		}
	case "txt":
		return net.LookupTXT(target)
	}
	return ret, nil
}

func doPollingNTP(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		setPollingError("ntp", p, fmt.Errorf("Node not dond"))
		return
	}
	lr := make(map[string]string)
	ok = false
	for i := 0; !ok && i <= p.Retry; i++ {
		options := ntp.QueryOptions{Timeout: time.Duration(p.Timeout) * time.Second}
		r, err := ntp.QueryWithOptions(n.IP, options)
		if err != nil {
			astiLogger.Debugf("doPollingNTP err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		p.LastVal = float64(r.RTT.Nanoseconds())
		lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
		lr["stratum"] = fmt.Sprintf("%d", r.Stratum)
		lr["refid"] = fmt.Sprintf("%d", r.ReferenceID)
		lr["offset"] = fmt.Sprintf("%d", r.ClockOffset.Nanoseconds())
		delete(lr, "error")
		ok = true
	}
	p.LastResult = makeLastResult(lr)
	if ok {
		setPollingState(p, "normal")
		return
	}
	setPollingState(p, p.Level)
	return
}

func setPollingError(s string, p *pollingEnt, err error) {
	astiLogger.Errorf("%s error Polling=%s err=%v", s, p.Polling, err)
	lr := make(map[string]string)
	lr["error"] = fmt.Sprintf("%v", err)
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "unknown")
}
