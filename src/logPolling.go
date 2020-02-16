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

	astilog "github.com/asticode/go-astilog"
)

type grokEnt struct {
	Pat string
	Ok  string
}

var (
	grokMap = map[string]*grokEnt{
		"EPSLOGIN":  &grokEnt{Pat: `Login %{GREEDYDATA:stat}: \[%{USER:user}\].+cli %{MAC:client}`, Ok: "OK"},
		"FZLOGIN":   &grokEnt{Pat: `FileZen: %{IP:client} %{USER:user} "Authentication %{GREEDYDATA:stat}`, Ok: "succeeded."},
		"NAOSLOGIN": &grokEnt{Pat: `Login %{GREEDYDATA:stat}: \[.+\] %{USER:user}`, Ok: "Success"},
	}
)

func loadGrokMap(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
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
	return nil
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
		astilog.Errorf("Invalid log watch format Polling=%s", p.Polling)
		p.LastResult = "Invalid log watch format"
		setPollingState(p, "unkown")
		return
	}
	astilog.Debugf("%q", cmds)
	filter := cmds[0]
	extractor := cmds[1]
	script := cmds[2]
	if _, err := regexp.Compile(filter); err != nil {
		astilog.Errorf("Invalid log watch format Polling=%s err=%v", p.Polling, err)
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
		astilog.Errorf("Invalid log watch format Polling=%s err=%v", p.Polling, err)
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
			astilog.Errorf("Invalid log watch format Polling=%s err=%v", p.Polling, err)
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
		astilog.Errorf("Invalid syslogpri watch format Polling=%s err=%v", p.Polling, err)
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

func doPollingSyslogUser(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	cmds := splitCmd(p.Polling)
	if len(cmds) != 2 {
		astilog.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		astilog.Errorf("Invalid SyslogUser format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		astilog.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
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
			astilog.Errorf("err=%v", err)
			continue
		}
		astilog.Infof("%v", values)
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
		astilog.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	astilog.Debugf("%q", cmds)
	filter := cmds[0]
	mode := cmds[1]
	if _, err := regexp.Compile(filter); err != nil {
		astilog.Errorf("Invalid SyslogUser format Polling=%s err=%v", p.Polling, err)
		p.LastResult = "Invalid SyslogUser format"
		setPollingState(p, "unkown")
		return
	}
	grokEnt, ok := grokMap[mode]
	if !ok {
		astilog.Errorf("Invalid SyslogUser format Polling=%s", p.Polling)
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
	for _, l := range logs {
		values, err := g.Parse(mode, string(l.Log))
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
		bytes, ok := values["bytes"]
		if !ok {
			continue
		}
		nProt := getProt(prot)
		nSPort, _ := strconv.Atoi(sport)
		nDPort, _ := strconv.Atoi(dport)
		nBytes, _ := strconv.Atoi(bytes)

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
