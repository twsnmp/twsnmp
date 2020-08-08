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
	return
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
		astiLogger.Errorf("Invalid log watch format Polling=%s", p.Polling)
		p.LastResult = "Invalid log watch format"
		setPollingState(p, "unkown")
		return
	}
	astiLogger.Debugf("%q", cmds)
	filter := cmds[0]
	extractor := cmds[1]
	script := cmds[2]
	if _, err := regexp.Compile(filter); err != nil {
		astiLogger.Errorf("Invalid log watch format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid log watch format"
		setPollingState(p, "unkown")
		return
	}
	vm := otto.New()
	lr := make(map[string]string)
	json.Unmarshal([]byte(p.LastResult), &lr)
	st := lr["lastTime"]
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
	vm.Set("count", len(logs))
	vm.Set("interval", p.PollInt)
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
		astiLogger.Errorf("Invalid log watch format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid log watch format"
		setPollingState(p, "unkown")
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})

	for _, l := range logs {
		values, err := g.Parse(extractor, string(l.Log))
		if err != nil {
			continue
		}
		for k, v := range values {
			vm.Set(k, v)
			lr[k] = v
		}
		value, err := vm.Run(script)
		if err == nil {
			p.LastResult = makeLastResult(lr)
			if lv, err := vm.Get("LastVal"); err == nil {
				if lvf, err := lv.ToFloat(); err == nil {
					p.LastVal = lvf
				}
			}
			if ok, _ := value.ToBoolean(); !ok {
				setPollingState(p, p.Level)
				return
			}
		} else {
			astiLogger.Errorf("Invalid log watch format Polling=%s err=%v", p.Polling, err)
			p.LastResult = "Invalid log watch format"
			setPollingState(p, "unkown")
			return
		}
	}
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "normal")
	return
}

var syslogPriFilter = regexp.MustCompile(`"priority":(\d+),`)

func doPollingSyslogPri(p *pollingEnt) bool {
	_, err := regexp.Compile(p.Polling)
	if err != nil {
		astiLogger.Errorf("Invalid syslogpri watch format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid syslogpri watch format"
		setPollingState(p, "unkown")
		updatePolling(p)
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
	p.LastResult = ""
	for pri, c := range priMap {
		if p.LastResult != "" {
			p.LastResult += ";"
		}
		p.LastResult += fmt.Sprintf("%d=%d", pri, c)
	}
	setPollingState(p, "normal")
	updatePolling(p)
	return true
}

func doPollingSyslogDevice(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) != 2 {
		astiLogger.Errorf("Invalid SyslogDevice format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogDevice format"
		setPollingState(p, "unkown")
		return
	}
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		astiLogger.Errorf("Invalid SyslogDevice format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid SyslogDevice format"
		setPollingState(p, "unkown")
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		astiLogger.Errorf("Invalid SyslogDevice format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogDevice format"
		setPollingState(p, "unkown")
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	g.AddPattern(mode, grokEnt.Pat)
	lr := make(map[string]string)
	json.Unmarshal([]byte(p.LastResult), &lr)
	st := lr["lastTime"]
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
	return
}

func doPollingSyslogUser(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	cmds := splitCmd(p.Polling)
	if len(cmds) != 2 {
		astiLogger.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		astiLogger.Errorf("Invalid SyslogUser format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		astiLogger.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	g.AddPattern(mode, grokEnt.Pat)
	lr := make(map[string]string)
	json.Unmarshal([]byte(p.LastResult), &lr)
	st := lr["lastTime"]
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
		client, _ := values["client"]
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
	return
}

func doPollingSyslogFlow(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) != 2 {
		astiLogger.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		astiLogger.Errorf("Invalid SyslogUser format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		astiLogger.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	g.AddPattern(mode, grokEnt.Pat)

	lr := make(map[string]string)
	json.Unmarshal([]byte(p.LastResult), &lr)
	st := lr["lastTime"]
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
	return
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
