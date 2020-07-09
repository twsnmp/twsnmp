package main

// cmdPolling.go : 外部コマンド実行で監視する。

import (
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/robertkrimen/otto"
	"github.com/vjeantet/grok"
	"golang.org/x/crypto/ssh"
)

func setPollingError(s string, p *pollingEnt, err error) {
	astiLogger.Errorf("%s error Polling=%s err=%v", s, p.Polling, err)
	p.LastResult = fmt.Sprintf("err=%v", err)
	setPollingState(p, "unkown")
}

func doPollingCmd(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) < 3 {
		setPollingError("cmd", p, fmt.Errorf("No Cmd"))
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
		setPollingError("cmd", p, fmt.Errorf("No Cmd"))
		return
	}
	tio := &timeout.Timeout{
		Cmd:       exec.Command(cl[0], cl[1:]...),
		Duration:  time.Duration(p.Timeout) * time.Second,
		KillAfter: 5 * time.Second,
	}
	exitStatus, stdout, stderr, err := tio.Run()
	if err != nil {
		setPollingError("cmd", p, err)
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
			setPollingError("cmd", p, err)
			return
		}
		for k, v := range values {
			vm.Set(k, v)
			lr[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("cmd", p, err)
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

func doPollingSSH(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) < 3 {
		setPollingError("ssh", p, fmt.Errorf("No Cmd"))
		return
	}
	cmd := cmds[0]
	extractor := cmds[1]
	script := cmds[2]
	port := "22"
	if len(cmds) > 3 {
		port = cmds[3]
	}
	vm := otto.New()
	lr := make(map[string]string)
	cl := strings.Split(cmd, " ")
	if len(cl) < 1 {
		setPollingError("ssh", p, fmt.Errorf("No Cmd"))
		return
	}
	client, session, err := sshConnectToHost(p, port)
	if err != nil {
		astiLogger.Errorf("ssh error Polling=%s err=%v", p.Polling, err)
		p.LastResult = fmt.Sprintf("err=%v", err)
		p.LastVal = -1.0
		setPollingState(p, p.Level)
		return
	}
	defer func() {
		session.Close()
		client.Close()
	}()
	out, err := session.CombinedOutput(cmd)
	if err != nil {
		if e, ok := err.(*ssh.ExitError); ok {
			p.LastVal = float64(e.Waitmsg.ExitStatus())
		} else {
			astiLogger.Errorf("ssh error Polling=%s err=%v", p.Polling, err)
			p.LastResult = fmt.Sprintf("err=%v", err)
			p.LastVal = -2.0
			setPollingState(p, p.Level)
			return
		}
	} else {
		p.LastVal = 0.0
	}
	lr["lastTime"] = time.Now().Format("2006-01-02T15:04")
	lr["exitCode"] = fmt.Sprintf("%d", int(p.LastVal))
	vm.Set("interval", p.PollInt)
	vm.Set("exitCode", int(p.LastVal))
	if extractor != "" {
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		values, err := g.Parse(extractor, string(out))
		if err != nil {
			setPollingError("ssh", p, err)
			return
		}
		for k, v := range values {
			vm.Set(k, v)
			lr[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("ssh", p, err)
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

func sshConnectToHost(p *pollingEnt, port string) (*ssh.Client, *ssh.Session, error) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return nil, nil, fmt.Errorf("node not found nodeID=%s", p.NodeID)
	}
	signer, err := ssh.ParsePrivateKey([]byte(getRawKeyPem(mapConf.PrivateKey)))
	if err != nil {
		astiLogger.Errorf("sshConnectToHost err=%v", err)
		return nil, nil, fmt.Errorf("No PrivateKey for SSH")
	}
	sshConfig := &ssh.ClientConfig{
		User:    n.User,
		Auth:    []ssh.AuthMethod{},
		Timeout: time.Duration(p.Timeout) * time.Second,
	}
	sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	if n.Password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(n.Password))
	}
	if n.PublicKey != "" {
		pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(n.PublicKey))
		if err != nil {
			return nil, nil, fmt.Errorf("Invalid PublicKey=%s", n.PublicKey)
		}
		sshConfig.HostKeyCallback = ssh.FixedHostKey(pubkey)
	} else {
		sshConfig.HostKeyCallback =
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				n.PublicKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
				updateNode(n)
				pollingStateChangeCh <- p
				return nil
			}
		//ssh.InsecureIgnoreHostKey()
	}
	conn, err := net.DialTimeout("tcp", n.IP+":"+port, time.Duration(p.Timeout)*time.Second)
	if err != nil {
		return nil, nil, err
	}
	conn.SetDeadline(time.Now().Add(time.Second * time.Duration(p.PollInt-5)))
	c, ch, req, err := ssh.NewClientConn(conn, n.IP+":"+port, sshConfig)
	if err != nil {
		return nil, nil, err
	}
	client := ssh.NewClient(c, ch, req)
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, nil, err
	}
	return client, session, nil
}
