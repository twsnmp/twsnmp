package main

// logPolling.go : LOGの監視を行う。

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/vjeantet/grok"

	astilog "github.com/asticode/go-astilog"
)

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
