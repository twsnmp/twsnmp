package main

// logPolling.go : LOGの監視を行う。

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/vjeantet/grok"
)

type grokEnt struct {
	Pat string
	Ok  string
}

var (
	grokMap = map[string]*grokEnt{
		"EPSLOGIN":    {Pat: `Login %{GREEDYDATA:stat}: \[%{USER:user}\].+cli %{MAC:client}`, Ok: "OK"},
		"FZLOGIN":     {Pat: `FileZen: %{IP:client} %{USER:user} "Authentication %{GREEDYDATA:stat}`, Ok: "succeeded."},
		"NAOSLOGIN":   {Pat: `Login %{GREEDYDATA:stat}: \[.+\] %{USER:user}`, Ok: "Success"},
		"LAPDEVICE":   {Pat: `mac=%{MAC:mac} ip=%{IP:ip}`},
		"WELFFLOW":    {Pat: `src=%{IP:src}:%{:sport}:.+ dst=%{IP:dst}:%{BASE10NUM:dport}:.+proto=%{WORD:prot}/.+ sent=%{BASE10NUM:sent} .+rcvd=%{BASE10NUM:rcvd}`},
		"OPENWEATHER": {Pat: `"weather":.+"main":\s*"%{WORD:weather}".+"main":.+"temp":\s*%{BASE10NUM:temp}.+"feels_like":\s*%{BASE10NUM:feels_like}.+"temp_min":\s*%{BASE10NUM:temp_min}.+"temp_max":\s*%{BASE10NUM:temp_max}.+"pressure":\s*%{BASE10NUM:pressure}.+"humidity":\s*%{BASE10NUM:humidity}.+"wind":\s*{"speed":\s*%{BASE10NUM:wind}`},
		"UPTIME":      {Pat: `load average: %{BASE10NUM:load1m}, %{BASE10NUM:load5m}, %{BASE10NUM:load15m}`},
	}
)

func loadGrokMap() {
	if mapConf.GrokPath == "" {
		return
	}
	f, err := os.Open(mapConf.GrokPath)
	if err != nil {
		astiLogger.Errorf("loadGrokMap err=%v", err)
		return
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.TrimSpace(s.Text())
		if len(l) < 1 || strings.HasPrefix(l, "#") {
			continue
		}
		e := splitCmd(l)
		if len(e) < 3 {
			continue
		}
		grokMap[e[0]] = &grokEnt{
			Pat: e[1],
			Ok:  e[2],
		}
	}
}

func splitCmd(p string) []string {
	ret := []string{}
	bInQ := false
	tmp := ""
	for _, c := range p {
		if c == '|' {
			if !bInQ {
				ret = append(ret, strings.TrimSpace(tmp))
				tmp = ""
			}
			continue
		}
		if c == '`' {
			bInQ = !bInQ
		} else {
			tmp += string(c)
		}
	}
	ret = append(ret, strings.TrimSpace(tmp))
	return ret
}

func makeLastResult(lr map[string]string) string {
	if js, err := json.Marshal(lr); err == nil {
		return string(js)
	}
	return ""
}

func doPollingLog(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) != 3 {
		setPollingError("log", p, fmt.Errorf("invalid log watch format"))
		return
	}
	filter := "`" + cmds[0] + "`"
	extractor := cmds[1]
	script := cmds[2]
	if _, err := regexp.Compile(filter); err != nil {
		setPollingError("log", p, fmt.Errorf("invalid log watch format"))
		return
	}
	vm := otto.New()
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(p.LastResult), &lr); err != nil {
		astiLogger.Errorf("doPollingLog err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(p.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := getLogs(&filterEnt{
		Filter:    filter,
		StartTime: st,
		EndTime:   et,
		LogType:   p.Type,
	})
	lr["lastTime"] = et
	_ = vm.Set("count", len(logs))
	_ = vm.Set("interval", p.PollInt)
	lr["count"] = fmt.Sprintf("%d", len(logs))
	p.LastVal = float64(len(logs))
	if extractor == "" {
		value, err := vm.Run(script)
		if err == nil {
			p.LastResult = makeLastResult(lr)
			if ok, _ := value.ToBoolean(); ok {
				setPollingState(p, "normal")
			} else {
				setPollingState(p, p.Level)
			}
			return
		}
		setPollingError("log", p, fmt.Errorf("invalid log watch format"))
		return
	}
	grokEnt, ok := grokMap[extractor]
	if !ok {
		setPollingError("log", p, fmt.Errorf("no grok pattern"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
		setPollingError("log", p, fmt.Errorf("no grok pattern"))
		return
	}
	cap := fmt.Sprintf("%%{%s}", extractor)
	for _, l := range logs {
		values, err := g.Parse(cap, string(l.Log))
		if err != nil {
			continue
		}
		for k, v := range values {
			_ = vm.Set(k, v)
			lr[k] = v
		}
		value, err := vm.Run(script)
		if err == nil {
			if ok, _ := value.ToBoolean(); !ok {
				p.LastResult = makeLastResult(lr)
				setPollingState(p, p.Level)
				return
			}
		} else {
			setPollingError("log", p, fmt.Errorf("invalid log watch format"))
			return
		}
	}
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "normal")
}

var syslogPriFilter = regexp.MustCompile(`"priority":(\d+),`)

func doPollingSyslogPri(p *pollingEnt) bool {
	_, err := regexp.Compile(p.Polling)
	if err != nil {
		setPollingError("log", p, fmt.Errorf("invalid syslogpri watch format"))
		_ = updatePolling(p)
		return false
	}
	endTime := time.Unix((time.Now().Unix()/3600)*3600, 0)
	startTime := endTime.Add(-time.Hour * 1)
	if int64(p.LastVal) >= startTime.UnixNano() {
		// Skip
		return false
	}
	p.LastVal = float64(startTime.UnixNano())
	st := startTime.Format("2006-01-02T15:04")
	et := endTime.Format("2006-01-02T15:04")
	logs := getLogs(&filterEnt{
		Filter:    p.Polling,
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	priMap := make(map[int]int)
	for _, l := range logs {
		pa := syslogPriFilter.FindAllStringSubmatch(string(l.Log), -1)
		if pa == nil || len(pa) < 1 || len(pa[0]) < 2 {
			continue
		}
		pri, err := strconv.ParseInt(pa[0][1], 10, 64)
		if err != nil || pri < 0 || pri > 256 {
			continue
		}
		priMap[int(pri)]++
	}
	lr := make(map[string]string)
	for pri, c := range priMap {
		lr[fmt.Sprintf("pri_%d", pri)] = fmt.Sprintf("%d", c)
	}
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "normal")
	_ = updatePolling(p)
	return true
}

func doPollingSyslogDevice(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) != 2 {
		setPollingError("log", p, fmt.Errorf("invalid syslog device watch format"))
		return
	}
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		setPollingError("log", p, fmt.Errorf("invalid syslog device watch format"))
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		setPollingError("log", p, fmt.Errorf("invalid syslog device watch format"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(mode, grokEnt.Pat); err != nil {
		setPollingError("log", p, fmt.Errorf("invalid syslog device watch format err=%v", err))
		return
	}
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(p.LastResult), &lr); err != nil {
		astiLogger.Errorf("doPollingSyslogDevice err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(p.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := getLogs(&filterEnt{
		Filter:    filter,
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	lr["lastTime"] = et
	lr["count"] = fmt.Sprintf("%d", len(logs))
	count := 0
	cap := fmt.Sprintf("%%{%s}", mode)
	for _, l := range logs {
		values, err := g.Parse(cap, string(l.Log))
		if err != nil {
			astiLogger.Errorf("err=%v", err)
			continue
		}
		mac, ok := values["mac"]
		if !ok {
			continue
		}
		ip, ok := values["ip"]
		if !ok {
			continue
		}
		mac = normMACAddr(mac)
		count++
		deviceReportCh <- &deviceReportEnt{
			Time: l.Time,
			MAC:  mac,
			IP:   ip,
		}
	}
	p.LastVal = float64(count)
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "normal")
}

func doPollingSyslogUser(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	cmds := splitCmd(p.Polling)
	if len(cmds) != 2 {
		setPollingError("log", p, fmt.Errorf("invalid syslog user watch format"))
		return
	}
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		setPollingError("log", p, fmt.Errorf("invalid filter for syslog user"))
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		setPollingError("log", p, fmt.Errorf("invalid grok for syslog user"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(mode, grokEnt.Pat); err != nil {
		astiLogger.Errorf("doPollingSyslogUser err=%v", err)
	}
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(p.LastResult), &lr); err != nil {
		astiLogger.Errorf("doPollingSyslogUser err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(p.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := getLogs(&filterEnt{
		Filter:    filter,
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	lr["lastTime"] = et
	lr["count"] = fmt.Sprintf("%d", len(logs))
	okCount := 0
	totalCount := 0
	cap := fmt.Sprintf("%%{%s}", mode)
	for _, l := range logs {
		values, err := g.Parse(cap, string(l.Log))
		if err != nil {
			astiLogger.Errorf("err=%v", err)
			continue
		}
		stat, ok := values["stat"]
		if !ok {
			continue
		}
		user, ok := values["user"]
		if !ok {
			continue
		}
		client := values["client"]
		ok = grokEnt.Ok == stat
		totalCount++
		if ok {
			okCount++
		}
		userReportCh <- &userReportEnt{
			Time:   l.Time,
			Server: n.IP,
			Client: client,
			UserID: user,
			Ok:     ok,
		}
	}
	if totalCount > 0 {
		p.LastVal = float64(okCount) / float64(totalCount)
	} else {
		p.LastVal = 1.0
	}
	lr["total"] = fmt.Sprintf("%d", totalCount)
	lr["ok"] = fmt.Sprintf("%d", okCount)
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "normal")
}

func doPollingSyslogFlow(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) != 2 {
		setPollingError("syslogFlow", p, fmt.Errorf("invalid watch format"))
		return
	}
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		setPollingError("syslogFlow", p, fmt.Errorf("invalid filter"))
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		setPollingError("syslogFlow", p, fmt.Errorf("invalid grok"))
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	if err := g.AddPattern(mode, grokEnt.Pat); err != nil {
		setPollingError("syslogFlow", p, fmt.Errorf("invalid grok"))
		return
	}
	lr := make(map[string]string)
	st := ""
	if err := json.Unmarshal([]byte(p.LastResult), &lr); err != nil {
		astiLogger.Errorf("doPollingSyslogFlow err=%v", err)
	} else {
		st = lr["lastTime"]
	}
	if _, err := time.Parse("2006-01-02T15:04", st); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(p.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := getLogs(&filterEnt{
		Filter:    filter,
		StartTime: st,
		EndTime:   et,
		LogType:   "syslog",
	})
	lr["lastTime"] = et
	lr["count"] = fmt.Sprintf("%d", len(logs))
	count := 0
	cap := fmt.Sprintf("%%{%s}", mode)
	for _, l := range logs {
		values, err := g.Parse(cap, string(l.Log))
		if err != nil {
			continue
		}
		src, ok := values["src"]
		if !ok {
			continue
		}
		dst, ok := values["dst"]
		if !ok {
			continue
		}
		sport, ok := values["sport"]
		if !ok {
			continue
		}
		dport, ok := values["dport"]
		if !ok {
			continue
		}
		prot, ok := values["prot"]
		if !ok {
			continue
		}
		nBytes := 0
		for _, b := range []string{"bytes", "sent", "rcvd"} {
			bytes, ok := values[b]
			if ok {
				nB, _ := strconv.Atoi(bytes)
				nBytes += nB
			}
		}
		nProt := getProt(prot)
		nSPort, _ := strconv.Atoi(sport)
		nDPort, _ := strconv.Atoi(dport)
		flowReportCh <- &flowReportEnt{
			Time:    l.Time,
			SrcIP:   src,
			SrcPort: nSPort,
			DstIP:   dst,
			DstPort: nDPort,
			Prot:    nProt,
			Bytes:   int64(nBytes),
		}
	}
	p.LastVal = float64(count)
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "normal")
}

func getProt(p string) int {
	if strings.Contains(p, "tcp") {
		return 6
	}
	if strings.Contains(p, "udp") {
		return 17
	}
	if strings.Contains(p, "icmp") {
		return 1
	}
	return 0
}
