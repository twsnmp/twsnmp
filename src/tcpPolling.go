package main

// tcpPolling.go :TCP/HTTP(S)/TLSのポーリングを行う。

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	astilog "github.com/asticode/go-astilog"
)

func doPollingTCP(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	ok = false
	var rTime int64
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		conn, err := net.DialTimeout("tcp", n.IP+":"+p.Polling, time.Duration(p.Timeout)*time.Second)
		endTime := time.Now().UnixNano()
		if err != nil {
			astilog.Debugf("doPollingTCP err=%v", err)
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
	_, ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s", p.NodeID)
		return
	}
	ok = false
	var rTime int64
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		err := doHTTPGet(p)
		endTime := time.Now().UnixNano()
		if err != nil {
			astilog.Debugf("doPollingHTTP err=%v", err)
			p.LastResult = fmt.Sprintf("%v", err)
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		setPollingState(p, "normal")
	} else {
		setPollingState(p, p.Level)
	}
	updatePolling(p)
}

var insecureTransport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var insecureClient = &http.Client{Transport: insecureTransport}

func doHTTPGet(p *pollingEnt) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Timeout)*time.Second)
	defer cancel()
	req, err := http.NewRequest(http.MethodGet, p.Polling, nil)
	if err != nil {
		return err
	}
	if p.Type == "https" {
		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		p.LastResult = resp.Status
		return nil
	}
	resp, err := insecureClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	p.LastResult = resp.Status
	return nil
}

func doPollingTLS(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s", p.NodeID)
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
			astilog.Debugf("doPollingTLS err=%v", err)
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
		astilog.Errorf("loadTLSParamsMap err=%v", err)
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
