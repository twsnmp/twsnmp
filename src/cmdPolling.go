package main

// cmdPolling.go : 外部コマンド実行で監視する。

import (
	"fmt"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Songmu/timeout"
	"github.com/robertkrimen/otto"
	"github.com/vjeantet/grok"
	"golang.org/x/crypto/ssh"
)

func doPollingCmd(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) < 3 {
		setPollingError("cmd", p, fmt.Errorf("no cmd"))
		return
	}
	cmd := cmds[0]
	extractor := cmds[1]
	script := cmds[2]
	vm := otto.New()
	lr := make(map[string]string)
	cl := strings.Split(cmd, " ")
	if len(cl) < 1 {
		setPollingError("cmd", p, fmt.Errorf("no cmd"))
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
	lr["stderr"] = stderr
	lr["exitCode"] = fmt.Sprintf("%d", exitStatus.Code)
	if err := vm.Set("exitCode", exitStatus.Code); err != nil {
		astiLogger.Errorf("doPollingCmd err=%v", err)
	}
	if err := vm.Set("interval", p.PollInt); err != nil {
		astiLogger.Errorf("doPollingCmd err=%v", err)
	}
	p.LastVal = float64(exitStatus.Code)
	if extractor != "" {
		grokEnt, ok := grokMap[extractor]
		if !ok {
			astiLogger.Errorf("No grok pattern Polling=%s", p.Polling)
			setPollingError("cmd", p, fmt.Errorf("no grok pattern"))
			return
		}
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		if err := g.AddPattern(extractor, grokEnt.Pat); err != nil {
			astiLogger.Errorf("doPollingCmd err=%v", err)
		}
		cap := fmt.Sprintf("%%{%s}", extractor)
		values, err := g.Parse(cap, string(stdout))
		if err != nil {
			setPollingError("cmd", p, err)
			return
		}
		for k, v := range values {
			if err := vm.Set(k, v); err != nil {
				astiLogger.Errorf("doPollingCmd err=%v", err)
			}
			lr[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("cmd", p, err)
		return
	}
	p.LastVal = 0.0
	for k, v := range lr {
		if strings.Contains(script, k) {
			if fv, err := strconv.ParseFloat(v, 64); err != nil || !math.IsNaN(fv) {
				p.LastVal = fv
			}
			break
		}
	}
	p.LastResult = makeLastResult(lr)
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(p, "normal")
		return
	}
	setPollingState(p, p.Level)
}

func doPollingSSH(p *pollingEnt) {
	cmds := splitCmd(p.Polling)
	if len(cmds) < 3 {
		setPollingError("ssh", p, fmt.Errorf("no cmd"))
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
		setPollingError("ssh", p, fmt.Errorf("no cmd"))
		return
	}
	client, session, err := sshConnectToHost(p, port)
	if err != nil {
		astiLogger.Errorf("ssh error Polling=%s err=%v", p.Polling, err)
		lr["error"] = fmt.Sprintf("%v", err)
		p.LastResult = makeLastResult(lr)
		p.LastVal = 0.0
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
			lr["error"] = fmt.Sprintf("%v", err)
			p.LastResult = makeLastResult(lr)
			p.LastVal = 0.0
			setPollingState(p, p.Level)
			return
		}
	} else {
		p.LastVal = 0.0
	}
	lr["lastTime"] = time.Now().Format("2006-01-02T15:04")
	lr["exitCode"] = fmt.Sprintf("%d", int(p.LastVal))
	_ = vm.Set("interval", p.PollInt)
	_ = vm.Set("exitCode", int(p.LastVal))
	if extractor != "" {
		grokEnt, ok := grokMap[extractor]
		if !ok {
			astiLogger.Errorf("No grok pattern Polling=%s", p.Polling)
			setPollingError("ssh", p, fmt.Errorf("no grok pattern"))
			return
		}
		g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
		_ = g.AddPattern(extractor, grokEnt.Pat)
		cap := fmt.Sprintf("%%{%s}", extractor)
		values, err := g.Parse(cap, string(out))
		if err != nil {
			setPollingError("ssh", p, err)
			return
		}
		for k, v := range values {
			_ = vm.Set(k, v)
			lr[k] = v
		}
	}
	value, err := vm.Run(script)
	if err != nil {
		setPollingError("ssh", p, err)
		return
	}
	p.LastVal = 0.0
	for k, v := range lr {
		if strings.Contains(script, k) {
			if fv, err := strconv.ParseFloat(v, 64); err != nil || !math.IsNaN(fv) {
				p.LastVal = fv
			}
			break
		}
	}
	p.LastResult = makeLastResult(lr)
	if ok, _ := value.ToBoolean(); ok {
		setPollingState(p, "normal")
		return
	}
	setPollingState(p, p.Level)
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
		return nil, nil, fmt.Errorf("no private key for ssh")
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
			return nil, nil, fmt.Errorf("invalid public key=%s", n.PublicKey)
		}
		sshConfig.HostKeyCallback = ssh.FixedHostKey(pubkey)
	} else {
		sshConfig.HostKeyCallback =
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				n.PublicKey = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
				if err := updateNode(n); err != nil {
					astiLogger.Errorf("sshConnectToHost err=%v", err)
				}
				pollingStateChangeCh <- p
				return nil
			}
		//ssh.InsecureIgnoreHostKey()
	}
	conn, err := net.DialTimeout("tcp", n.IP+":"+port, time.Duration(p.Timeout)*time.Second)
	if err != nil {
		return nil, nil, err
	}
	if err := conn.SetDeadline(time.Now().Add(time.Second * time.Duration(p.PollInt-5))); err != nil {
		astiLogger.Errorf("sshConnectToHost err=%v", err)
	}
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
