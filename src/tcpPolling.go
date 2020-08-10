package main

// tcpPolling.go :TCP/HTTP(S)/TLSのポーリングを行う。

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/vjeantet/grok"
)

func doPollingTCP(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	ok = false
	var rTime int64
	lr := make(map[string]string)
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := net.DialTimeout("tcp", n.IP+":"+p.Polling, time.Duration(p.Timeout)*time.Second)
		endTime := time.Now().UnixNano()
		if err != nil {
			astiLogger.Debugf("doPollingTCP err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		delete(lr, "error")
		lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	p.LastResult = makeLastResult(lr)
}

func doPollingHTTP(p *pollingEnt) {
	var ok bool
	var err error
	_, ok = nodes[p.NodeID]
	if !ok {
		setPollingError("http", p, fmt.Errorf("Node not found"))
		return
	}
	cmd := splitCmd(p.Polling)
	if len(cmd) < 1 {
		setPollingError("http", p, fmt.Errorf("No URL"))
		return
	}
	url := cmd[0]
	ok = false
	var rTime int64
	body := ""
	status := ""
	code := 0
	lr := make(map[string]string)
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		status, body, code, err = doHTTPGet(p, url)
		endTime := time.Now().UnixNano()
		if err != nil {
			astiLogger.Debugf("doPollingHTTP err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if len(cmd) > 2 {
		ok, lr, err = checkHTTPResp(p, cmd[1], cmd[2], status, body, code)
		if err != nil {
			setPollingError("http", p, err)
			return
		}
	} else {
		lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
		lr["status"] = status
	}
	if ok {
		delete(lr, "error")
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	p.LastResult = makeLastResult(lr)
}

func checkHTTPResp(p *pollingEnt, extractor, script, status, body string, code int) (bool, map[string]string, error) {
	lr := make(map[string]string)
	vm := otto.New()
	lr["status"] = status
	lr["code"] = fmt.Sprintf("%d", code)
	lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
	vm.Set("status", status)
	vm.Set("code", code)
	vm.Set("rtt", p.LastVal)
	if extractor == "" {
		value, err := vm.Run(script)
		if err != nil {
			return false, lr, err
		}
		p.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); ok {
			return true, lr, nil
		}
		return false, lr, nil
	}
	grokEnt, ok := grokMap[extractor]
	if !ok {
		return false, lr, fmt.Errorf("No grok pattern")
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	g.AddPattern(extractor, grokEnt.Pat)
	cap := fmt.Sprintf("%%{%s}", extractor)
	values, err := g.Parse(cap, body)
	if err != nil {
		return false, lr, err
	}
	for k, v := range values {
		vm.Set(k, v)
		lr[k] = v
	}
	value, err := vm.Run(script)
	if err != nil {
		return false, lr, err
	}
	if ok, _ := value.ToBoolean(); ok {
		return true, lr, nil
	}
	return false, lr, nil
}

var insecureTransport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var insecureClient = &http.Client{Transport: insecureTransport}

func doHTTPGet(p *pollingEnt, url string) (string, string, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Timeout)*time.Second)
	defer cancel()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", "", 0, err
	}
	body := make([]byte, 64*1024)
	if p.Type == "https" {
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return "", "", 0, err
		}
		defer resp.Body.Close()
		_, err = resp.Body.Read(body)
		if err == io.EOF {
			err = nil
		}
		return resp.Status, string(body), resp.StatusCode, err
	}
	resp, err := insecureClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	_, err = resp.Body.Read(body)
	if err == io.EOF {
		err = nil
	}
	return resp.Status, string(body), resp.StatusCode, err
}

func doPollingTLS(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		setPollingError("tls", p, fmt.Errorf("Node not found"))
		return
	}
	cmd := splitCmd(p.Polling)
	mode := ""
	target := n.Name + ":443"
	script := ""
	if len(cmd) > 2 {
		mode = cmd[0]
		target = cmd[1]
		script = cmd[2]
	}

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	switch mode {
	case "verify":
		conf.InsecureSkipVerify = false
	case "version":
		if strings.Contains(script, "1.0") {
			conf.MaxVersion = tls.VersionTLS10
		} else if strings.Contains(script, "1.1") {
			conf.MinVersion = tls.VersionTLS11
			conf.MaxVersion = tls.VersionTLS11
		} else if strings.Contains(script, "1.2") {
			conf.MinVersion = tls.VersionTLS12
			conf.MaxVersion = tls.VersionTLS12
		} else if strings.Contains(script, "1.3") {
			conf.MinVersion = tls.VersionTLS13
			conf.MaxVersion = tls.VersionTLS13
		}
	}

	d := &net.Dialer{
		Timeout: time.Duration(p.Timeout) * time.Second,
	}
	ok = false
	var rTime int64
	var cs tls.ConnectionState
	lr := make(map[string]string)
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := tls.DialWithDialer(d, "tcp", target, conf)
		endTime := time.Now().UnixNano()
		if err != nil {
			astiLogger.Debugf("doPollingTLS err=%v", err)
			lr["error"] = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		cs = conn.ConnectionState()
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		lr = getTLSConnectioStateInfo(n.Name, &cs)
		lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
		if mode == "expire" {
			var d int
			if _, err := fmt.Sscanf(script, "%d", &d); err != nil && d > 0 {
				cert := getServerCert(n.Name, &cs)
				if cert != nil {
					na := cert.NotAfter.Unix()
					ct := time.Now().AddDate(0, 0, d).Unix()
					if ct > na {
						ok = false
					}
				} else {
					ok = false
				}
			}
		}
	}
	p.LastResult = makeLastResult(lr)
	if (ok && !strings.Contains(script, "!")) || (!ok && strings.Contains(script, "!")) {
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
}

var tlsCSMap = make(map[string]string)

func loadTLSParamsMap(path string) {
	file, err := os.Open(path)
	if err != nil {
		astiLogger.Errorf("loadTLSParamsMap err=%v", err)
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	var line []string
	for {
		line, err = reader.Read()
		if err != nil {
			break
		}
		if len(line) < 2 {
			continue
		}
		id := strings.Replace(line[0], ",", "", 1)
		id = strings.Replace(id, "0x", "", 2)
		id = strings.ToLower(id)
		name := line[1]
		if strings.HasPrefix(name, "TLS_") {
			tlsCSMap[id] = name
		}
	}
}

func getServerCert(host string, cs *tls.ConnectionState) *x509.Certificate {
	for _, cl := range cs.VerifiedChains {
		for _, c := range cl {
			if c.VerifyHostname(host) == nil {
				return c
			}
		}
	}
	for _, c := range cs.PeerCertificates {
		if c.VerifyHostname(host) == nil {
			return c
		}
	}
	return nil
}

func getTLSConnectioStateInfo(host string, cs *tls.ConnectionState) map[string]string {
	ret := make(map[string]string)
	switch cs.Version {
	case tls.VersionSSL30:
		ret["version"] = "SSLv3"
	case tls.VersionTLS10:
		ret["version"] = "TLSv1.0"
	case tls.VersionTLS11:
		ret["version"] = "TLSv1.1"
	case tls.VersionTLS12:
		ret["version"] = "TLSv1.2"
	case tls.VersionTLS13:
		ret["version"] = "TLSv1.3"
	default:
		ret["version"] = "Unknown"
	}
	id := fmt.Sprintf("%04x", cs.CipherSuite)
	if n, ok := tlsCSMap[id]; ok {
		ret["cipherSuite"] = n
	} else {
		ret["cipherSuite"] = id
	}
	if len(cs.VerifiedChains) > 0 {
		ret["valid"] = "true"
	} else {
		ret["valid"] = "false"
	}
	if cert := getServerCert(host, cs); cert != nil {
		ret["issuer"] = cert.Issuer.String()
		ret["subject"] = cert.Subject.String()
		ret["notAfter"] = cert.NotAfter.Format("2006/01/02")
		ret["subjectKeyID"] = fmt.Sprintf("%x", cert.SubjectKeyId)
	}
	return ret
}
