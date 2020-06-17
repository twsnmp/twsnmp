package main

// cmdPolling.go : 外部コマンド実行で監視する。

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/robertkrimen/otto"
	"github.com/vjeantet/grok"
)

func setCmdPollingError(p *pollingEnt, err error) {
	astiLogger.Errorf("Cmd polling error Polling=%s err=%v", p.Polling, err)
	p.LastResult = fmt.Sprintf("err=%v", err)
	setPollingState(p, "unkown")
}

func doPollingCmd(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) != 3 {
		setCmdPollingError(p, fmt.Errorf("No Cmd"))
		return
	}
	astiLogger.Debugf("%q", cmds)
	cmd := cmds[0]
	extractor := cmds[1]
	script := cmds[2]
	vm := otto.New()
	lr := make(map[string]string)
	cl := strings.Split(cmd, " ")
	if len(cl) < 1 {
		setCmdPollingError(p, fmt.Errorf("No Cmd"))
		return
	}
	tio := &timeout.Timeout{
		Cmd:       exec.Command(cl[0], cl[1:]...),
		Duration:  time.Duration(p.Timeout) * time.Second,
		KillAfter: 5 * time.Second,
	}
	exitStatus, stdout, stderr, err := tio.Run()
	if err != nil {
		setCmdPollingError(p, err)
		return
	}
	lr["lastTime"] = time.Now().Format("2006-01-02T15:04")
	// lr["stdout"] = stdout
	lr["stderr"] = stderr
	lr["exitCode"] = fmt.Sprintf("%d", exitStatus.Code)
	vm.Set("exitCode", exitStatus.Code)
	vm.Set("interval", p.PollInt)
	p.LastVal = float64(exitStatus.Code)
	if extractor != "" {
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		values, err := g.Parse(extractor, string(stdout))
		if err != nil {
			setCmdPollingError(p, err)
			return
		}
		for k, v := range values {
			vm.Set(k, v)
			lr[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setCmdPollingError(p, err)
		return
	}
	if lv, err := vm.Get("LastVal"); err == nil && lv.IsNumber() {
		if lvf, err := lv.ToFloat(); err == nil {
			p.LastVal = lvf
		}
	}
	p.LastResult = makeLastResult(lr)
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(p, "normal")
		return
	}
	setPollingState(p, p.Level)
	return
}
