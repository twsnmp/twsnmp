package main

// tcpPolling.go :TCP/HTTP(S)/TLSのポーリングを行う。

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"encoding/json"
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
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := net.DialTimeout("tcp", n.IP+":"+p.Polling, time.Duration(p.Timeout)*time.Second)
		endTime := time.Now().UnixNano()
		if err != nil {
			astiLogger.Debugf("doPollingTCP err=%v", err)
			p.LastResult = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		p.LastResult = ""
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	updatePolling(p)
}

func doPollingHTTP(p *pollingEnt) {
	var ok bool
	var err error
	_, ok = nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	cmd := splitCmd(p.Polling)
	if len(cmd) < 1 {
		astiLogger.Errorf("URL not found Polling=%s", p.Polling)
		return
	}
	url := cmd[0]
	ok = false
	var rTime int64
	body := ""
	status := ""
	code := 0
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		status, body, code, err = doHTTPGet(p, url)
		endTime := time.Now().UnixNano()
		if err != nil {
			astiLogger.Errorf("doPollingHTTP err=%v", err)
			p.LastResult = fmt.Sprintf("%v", err)
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if len(cmd) > 2 {
		ok = checkHTTPResp(p, cmd[1], cmd[2], status, body, code)
	} else {
		p.LastResult = status
	}
	if ok {
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	updatePolling(p)
}

func checkHTTPResp(p *pollingEnt, extractor, script, status, body string, code int) bool {
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
			astiLogger.Errorf("Invalid http get format Polling=%s err=%v", p.Polling, err)
			p.LastResult = fmt.Sprintf("vm.Run() err=%v", err)
			return false
		}
		p.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); ok {
			return true
		}
		return false
	}
	grokEnt, ok := grokMap[extractor]
	if !ok {
		p.LastResult = fmt.Sprintf("No grok pattern")
		astiLogger.Errorf("No grok pattern Polling=%s", p.Polling)
		return false
	}
	g, _ := grok.NewWithConfig(&grok.Config{NamedCapturesOnly: true})
	g.AddPattern(extractor, grokEnt.Pat)
	cap := fmt.Sprintf("%%{%s}", extractor)
	values, err := g.Parse(cap, body)
	if err != nil {
		astiLogger.Errorf("Invalid http get format Polling=%s err=%v", p.Polling, err)
		p.LastResult = fmt.Sprintf("g.Parse() err=%v", err)
		return false
	}
	for k, v := range values {
		vm.Set(k, v)
		lr[k] = v
	}
	value, err := vm.Run(script)
	if err != nil {
		astiLogger.Errorf("Invalid http get format Polling=%s err=%v", p.Polling, err)
		p.LastResult = fmt.Sprintf("vm.Run() err=%v", err)
		return false
	}
	p.LastResult = makeLastResult(lr)
	if lv, err := vm.Get("LastVal"); err == nil {
		if lvf, err := lv.ToFloat(); err == nil {
			p.LastVal = lvf
		}
	}
	if ok, _ := value.ToBoolean(); ok {
		return true
	}
	return false
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
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	if strings.Contains(p.Polling, "Verify") {
		conf.InsecureSkipVerify = false
	}
	if strings.Contains(p.Polling, "Version") {
		if strings.Contains(p.Polling, "1.0") {
			conf.MaxVersion = tls.VersionTLS10
		} else if strings.Contains(p.Polling, "1.1") {
			conf.MinVersion = tls.VersionTLS11
			conf.MaxVersion = tls.VersionTLS11
		} else if strings.Contains(p.Polling, "1.2") {
			conf.MinVersion = tls.VersionTLS12
			conf.MaxVersion = tls.VersionTLS12
		} else if strings.Contains(p.Polling, "1.3") {
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
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := tls.DialWithDialer(d, "tcp", n.IP+":"+p.Polling, conf)
		endTime := time.Now().UnixNano()
		if err != nil {
			astiLogger.Debugf("doPollingTLS err=%v", err)
			p.LastResult = fmt.Sprintf("%v", err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		cs = conn.ConnectionState()
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		p.LastResult = getTLSConnectioStateInfo(n.Name, &cs)
		if strings.Contains(p.Polling, "Expire") {
			var d int
			if _, err := fmt.Sscanf(p.Polling, "Expire %d", &d); err != nil && d > 0 {
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
	if (ok && !strings.Contains(p.Polling, "!")) || (!ok && strings.Contains(p.Polling, "!")) {
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	updatePolling(p)
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

func getTLSConnectioStateInfo(host string, cs *tls.ConnectionState) string {
	var tlsInfo = struct {
		Version      string
		CipherSuite  string
		NotAfter     time.Time
		Subject      string
		SubjectKeyID string
		Issuer       string
		Valid        bool
	}{}
	switch cs.Version {
	case tls.VersionSSL30:
		tlsInfo.Version = "SSLv3"
	case tls.VersionTLS10:
		tlsInfo.Version = "TLSv1.0"
	case tls.VersionTLS11:
		tlsInfo.Version = "TLSv1.1"
	case tls.VersionTLS12:
		tlsInfo.Version = "TLSv1.2"
	case tls.VersionTLS13:
		tlsInfo.Version = "TLSv1.3"
	default:
		tlsInfo.Version = "Unknown"
	}
	id := fmt.Sprintf("%04x", cs.CipherSuite)
	if n, ok := tlsCSMap[id]; ok {
		tlsInfo.CipherSuite = n
	} else {
		tlsInfo.CipherSuite = id
	}
	if len(cs.VerifiedChains) > 0 {
		tlsInfo.Valid = true
	}
	if cert := getServerCert(host, cs); cert != nil {
		tlsInfo.Issuer = cert.Issuer.String()
		tlsInfo.Subject = cert.Subject.String()
		tlsInfo.NotAfter = cert.NotAfter
		tlsInfo.SubjectKeyID = fmt.Sprintf("%x", cert.SubjectKeyId)
	}
	ret, err := json.Marshal(tlsInfo)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(ret)
}
