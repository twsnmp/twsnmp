package main

/*
polling.go :ポーリング処理を行う
ポーリングの種類は
(1)能動的なポーリング
 ping
 snmp - sysUptime,ifOperStatus,
 wget
（２）受動的なポーリング
 syslog
 trap
 netflow

*/

import (
	"context"
	"time"
	"strings"
	"fmt"
	"strconv"
	"regexp"
	"sort"
	"runtime"
	"net"
	"net/http"
	"crypto/tls"
	"encoding/csv"
	"os"
	gosnmp "github.com/soniah/gosnmp"

	astilog "github.com/asticode/go-astilog"

)

var (
	pollingStateChangeCh = make(chan *pollingEnt,10)
	doPollingCh = make(chan bool,10)
)

func pollingBackend(ctx context.Context) {
	go pingBackend(ctx)
	time.Sleep(time.Millisecond*100)
	var nextPoll int64
	for {
		select {
		case <-ctx.Done():
			return
		case  <- doPollingCh:
			{
				now := time.Now().UnixNano()
				if nextPoll > now {
					continue
				}
				nextPoll = now + (1000 * 1000 * 1000) * 2
				list := []*pollingEnt{}
				pollings.Range(func(_,v interface{}) bool {
					p := v.(*pollingEnt)
					if p.LastTime + (int64(p.PollInt) * 1000 * 1000 * 1000) < now {
						list = append(list,p)
					}
					return true
				})
				if len(list) < 1 {
					continue
				}
				astilog.Infof("New Polling %d NumGoroutine %d",len(list),runtime.NumGoroutine())
				sort.Slice(list,func (i,j int)bool {
					return list[i].LastTime < list[j].LastTime 
				})
				for i:=0; i < len(list);i++ {
					list[i].LastTime = time.Now().UnixNano()
					go doPolling(list[i])
					time.Sleep(time.Millisecond * 1)
				}
			}
		}
	}
}

func setPollingState(p *pollingEnt,newState string){
	sendEvent := false
	if newState == "normal" {
		if p.State != "normal" && p.State != "repair" {
			if p.State == "unkown"{
				p.State = "normal"
			} else {
				p.State = "repair"
			}
			sendEvent = true
		}
	} else if newState == "unkown" {
		if p.State != "unkown" {
			p.State = "unkown"
			sendEvent = true
		}
	}	else {
		if p.State != p.Level {
			p.State = p.Level
			sendEvent = true
		}
	}
	if sendEvent {
		nodeName := "Unknown"
		if n,ok := nodes[p.NodeID]; ok {
			nodeName = n.Name
		}
		pollingStateChangeCh <- p
		addEventLog(eventLogEnt{
			Type:"polling",
			Level: p.State,
			NodeID: p.NodeID,
			NodeName: nodeName,
			Event: fmt.Sprintf("ポーリング状態変化:%s(%s):%f:%s",p.Name,p.Type,p.LastVal,p.LastResult),
		})
	}
}

func doPolling(p *pollingEnt){
	oldState := p.State
	switch p.Type {
	case "ping":
		doPollingPing(p)
	case "snmp":
		doPollingSnmp(p)
	case "tcp":
		doPollingTCP(p)
	case "http","https":
		doPollingHTTP(p)
	case "tls":
		doPollingTLS(p)
	case "dns":
		doPollingDNS(p)
	case "syslog","trap","netflow","ipfix":
		doPollingLog(p)
		updatePolling(p)
	case "syslogpri":
		if !doPollingSyslogPri(p) {
			return
		}
	}
	if p.LogMode == 1 || p.LogMode == 3 || (p.LogMode == 2 && oldState != p.State) {
		if err := addPollingLog(p);err != nil {
			astilog.Errorf("addPollingLog err=%v",err)
		}
	}
}

func doPollingPing(p *pollingEnt){
	n,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	r := doPing(n.IP,p.Timeout,p.Retry,64)
	p.LastVal = float64(r.Time)
	if r.Stat == pingOK {
		p.LastResult = ""
		setPollingState(p,"normal")
	}	else {
		p.LastResult = fmt.Sprintf("%v",r.Error)
		setPollingState(p,p.Level)
	}
	updatePolling(p)
}

func doPollingSnmp(p *pollingEnt){
	n,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(p.Timeout) * time.Second,
		Retries:            p.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	err := agent.Connect()
	if err != nil {
		astilog.Errorf("SNMP agent.Connect err=%v",err)
		return
	}
	defer agent.Conn.Close()
	ps,mode := parseSnmpPolling(p.Polling)
	if ps == "" {
		astilog.Errorf("Empty SNMP Polling %s",p.Name)
		return
	}
	if ps == "sysUpTime" {
		doPollingSnmpSysUpTime(p,agent)
	} else if strings.HasPrefix(ps,"ifOperStatus.") {
		doPollingSnmpIF(p,ps,agent)
	} else {
		doPollingSnmpOther(p,ps,mode,agent)
	}
	updatePolling(p)
}

func parseSnmpPolling(s string) (string,string) {
	a :=  strings.Split(s,"|")
	if len(a) < 1 {
		return "",""
	}
	ps := strings.TrimSpace(a[0])
	if len(a) < 2 {
		return ps,""
	}
	mode := strings.TrimSpace(a[1])
	return ps,mode
}

func doPollingSnmpSysUpTime(p *pollingEnt,agent *gosnmp.GoSNMP){
	oids := []string{mib.NameToOID("sysUpTime.0")}
	result, err := agent.Get(oids)
	if err != nil {
		p.LastResult = fmt.Sprintf("%v",err)
		setPollingState(p,"unkown")
		return
	}
	var uptime int64
	for _, variable := range result.Variables {
		if variable.Name == mib.NameToOID("sysUpTime.0") {
			uptime = gosnmp.ToBigInt(variable.Value).Int64()
			break
		}
	}
	if uptime == 0 {
		p.LastResult = ""
		setPollingState(p,"unkown")
		return
	}
	p.LastVal = float64(uptime)
	if p.LastResult == "" {
		p.LastResult = fmt.Sprintf("%d",uptime)
		return
	}
	if lastUptime,err := strconv.ParseInt(p.LastResult,10,64);err != nil {
		p.LastResult = fmt.Sprintf("%d",uptime)
		setPollingState(p,"unkown")
	} else {
		p.LastResult = fmt.Sprintf("%d",uptime)
		if lastUptime < uptime {
			setPollingState(p,"normal")
			return
		}
		setPollingState(p,p.Level)
	}
}

func doPollingSnmpIF(p *pollingEnt,ps string,agent *gosnmp.GoSNMP) {
	a := strings.Split(ps,".")
	if len(a) < 2 {
		p.LastResult = "Invalid format"
		setPollingState(p,"unkown")
		return
	}
	oids := []string{mib.NameToOID("ifOperStatus."+a[1]),mib.NameToOID("ifAdminState."+a[1])}
	result, err := agent.Get(oids)
	if err != nil {
		p.LastResult = "Invalid MIB Name"
		setPollingState(p,"unkown")
		return
	}
	var oper int64
	var admin int64
	for _, variable := range result.Variables {
		if strings.HasPrefix(mib.OIDToName(variable.Name),"ifOperStatus") {
			oper = gosnmp.ToBigInt(variable.Value).Int64()
		} else if strings.HasPrefix(mib.OIDToName(variable.Name),"ifAdminStatus") {
			admin = gosnmp.ToBigInt(variable.Value).Int64()
		}
	}
	p.LastVal = float64(oper)
	p.LastResult = fmt.Sprintf("oper=%d;admin=%d",oper,admin)
	if oper == 1 {
		setPollingState(p,"normal")
		return
	} else if admin == 2 {
		setPollingState(p,"normal")
		return
	} else if oper == 2 && admin == 1 {
		setPollingState(p,p.Level)
		return
	}
	setPollingState(p,"unkown")
	return
}

func doPollingSnmpOther(p *pollingEnt,ps,mode string,agent *gosnmp.GoSNMP) {
	a := strings.Split(ps," ")
	if len(a) < 3 {
		p.LastResult = "Invalid format"
		setPollingState(p,"unkown")
		return
	}
	m := strings.TrimSpace(a[0])
	op := strings.TrimSpace(a[1])
	cv := strings.TrimSpace(a[2])
	oids := []string{mib.NameToOID(m)}
	if mode == "ps" {
		oids = append(oids,mib.NameToOID("sysUpTime.0"))
	}
	result, err := agent.Get(oids)
	if err != nil {
		p.LastResult = fmt.Sprintf("%v",err)
		setPollingState(p,"unkown")
		return
	}
	var iv int64
	var sut int64
	var sv string
	hitIv := false
	hitSv := false
	for _, variable := range result.Variables {
		if variable.Name == mib.NameToOID(ps) {
			if variable.Type == gosnmp.OctetString {
				sv = string(variable.Value.([]byte))
				hitSv = true
			} else if variable.Type == gosnmp.ObjectIdentifier {
				sv = mib.OIDToName(variable.Value.(string))
				hitSv = true
			} else {
				iv = gosnmp.ToBigInt(variable.Value).Int64()
				hitIv = true
			}
		} else if variable.Name == mib.NameToOID("sysUpTime.0"){
			sut = gosnmp.ToBigInt(variable.Value).Int64()
		}
	}
	if !hitIv && !hitSv {
		p.LastResult = "Invalid MIB"
		setPollingState(p,"unkown")
		return
	}
	if hitIv {
		sv = fmt.Sprintf("%d,%d",iv,sut)
	}
	if mode == "ps" || mode == "delta" {
		if !strings.Contains(p.LastResult,",") {
			p.LastResult =  sv
			return
		}
	}
	r := false
	if hitSv {
		switch op {
		case "=","==":
			r = sv == cv 
		case "~=":
			r = strings.Contains(sv,cv)
		case "<":
			r = strings.Compare(sv,cv) < 0
		case ">":
			r = strings.Compare(sv,cv) > 0
		default:
			p.LastResult = "Invalid Operator"
			setPollingState(p,"unkown")
			return
		}
		p.LastResult = sv 
	} else {
		civ,err :=  strconv.ParseInt(cv,10,64)
		if err != nil {
			p.LastResult = fmt.Sprintf("%d,%d",iv,sut)
			setPollingState(p,"unkown")
			return
		}
		var liv int64
		var lsut int64
		n,err :=  fmt.Sscanf(p.LastResult,"%d,%d",&liv,lsut)
		if err != nil || n != 2 {
			p.LastResult = fmt.Sprintf("%d,%d",iv,sut)
			setPollingState(p,"unkown")
			return
		}
		if mode == "ps" {
			dsut := sut -  lsut
			if dsut <= 0 {
				p.LastResult = fmt.Sprintf("%d,%d",iv,sut)
				setPollingState(p,"unkown")
				return
			}
			iv = (100*(iv-liv))/dsut
		} else if mode == "delta" {
			iv -= liv
		}
		switch op {
		case "=","==":
			r = iv == civ 
		case "!=":
			r = iv != civ 
		case "<":
			r =  iv < civ
		case ">":
			r = iv > civ
		case "<=":
			r =  iv <= civ
		case ">=":
			r = iv >= civ
		default:
			p.LastResult = "Invalid Operator"
			setPollingState(p,"unkown")
			return
		}
		p.LastVal = float64(iv)
	}
	if r {
		setPollingState(p,"normal")
		return
	}
	setPollingState(p,p.Level)
	return
}

var logPollRegex = regexp.MustCompile(`\s*(\S+.+\S+)\s*\|\s*(count|val)\s*(=|<|>|>=|<=|!=)\s*([-.0-9]+)`)

func doPollingLog(p *pollingEnt) {
	// 正規表現でログ検索定義を取得
	a := logPollRegex.FindAllStringSubmatch(p.Polling, -1)
	if  a == nil || len(a) != 1  {
		astilog.Errorf("Invalid log watch format Polling=%s",p.Polling)
		p.LastResult = "Invalid log watch format"
		setPollingState(p,"unkown")
		return
	}
	filter := a[0][1]
	f,err := regexp.Compile(filter)
	if err != nil {
		astilog.Errorf("Invalid log watch format Polling=%s err=%v",p.Polling,err)
		p.LastResult = "Invalid log watch format"
		setPollingState(p,"unkown")
		return
	}
	key    := a[0][2]
	op    := a[0][3]
	vs   := a[0][4]
	vc,err := strconv.ParseFloat(vs,64)
	if err != nil {
		astilog.Errorf("Invalid log watch format Polling=%s err=%v",p.Polling,err)
		p.LastResult = "Invalid log watch format"
		setPollingState(p,"unkown")
		return
	}
	st := p.LastResult
	if _,err := time.Parse("2006-01-02T15:04",p.LastResult); err != nil {
		st = time.Now().Add(-time.Second * time.Duration(p.PollInt)).Format("2006-01-02T15:04")
	}
	et := time.Now().Format("2006-01-02T15:04")
	logs := getLogs( &filterEnt {
		Filter: filter,
		StartTime: st,
		EndTime: et,
		LogType: p.Type,
	})
	p.LastResult = et
	if key == "count" {
		p.LastVal = float64(len(logs))
		if  cmpVal(op,p.LastVal,vc) {
			setPollingState(p,"normal")
		} else {
			setPollingState(p,p.Level)
		}
		return
	}
	bHit := false
	for _,l := range logs {
		va := f.FindAllStringSubmatch(string(l.Log),-1)
		if va == nil || len(va) < 1 || len(va[0]) < 2 {
			continue
		}
		vi,err := strconv.ParseFloat(va[0][1],64)
		if err != nil {
			continue
		}
		bHit = true
		p.LastVal = vi
		// ログの中に、一つでも異常がありれば、異常にする。
		if !cmpVal(op,vi,vc) {
			setPollingState(p,p.Level)
			return
		}
	}
	if bHit {
		setPollingState(p,"normal")
	} else {
		setPollingState(p,"unkown")
	}
	return
}

func cmpVal(op string,a,b float64) bool {
	switch op {
	case "=","==":
		return a == b
	case "!=":
		return a != b
	case "<":
		return  a < b
	case ">":
		return  a > b
	case "<=":
		return  a <= b
	case ">=":
		return a >= b
	default:
		return false
	}
}

var syslogPriFilter = regexp.MustCompile(`"priority":(\d+),`)

func doPollingSyslogPri(p *pollingEnt) bool {
	_,err := regexp.Compile(p.Polling)
	if err != nil {
		astilog.Errorf("Invalid syslogpri watch format Polling=%s err=%v",p.Polling,err)
		p.LastResult = "Invalid syslogpri watch format"
		setPollingState(p,"unkown")
		updatePolling(p)
		return false
	}
	endTime := time.Unix((time.Now().Unix()/3600)*3600,0)
	startTime := endTime.Add(-time.Hour * 1)
	if int64(p.LastVal) >= startTime.UnixNano() {
		// Skip
		return false
	}
	p.LastVal = float64(startTime.UnixNano())
	st := startTime.Format("2006-01-02T15:04")
	et := endTime.Format("2006-01-02T15:04")
	logs := getLogs( &filterEnt {
		Filter: p.Polling,
		StartTime: st,
		EndTime: et,
		LogType: "syslog",
	})
	priMap := make(map[int]int)
	for _,l := range logs {
		pa := syslogPriFilter.FindAllStringSubmatch(string(l.Log),-1)
		if pa == nil || len(pa) < 1 || len(pa[0]) < 2 {
			continue
		}
		pri,err := strconv.ParseInt(pa[0][1],10,64)
		if err != nil || pri < 0 || pri > 256 {
			continue
		}
		priMap[int(pri)]++
	}
	p.LastResult = ""
	for pri,c := range priMap{
		if p.LastResult != ""{
			p.LastResult += ";"
		}
		p.LastResult += fmt.Sprintf("%d=%d",pri,c)
	}
	setPollingState(p,"normal")
	updatePolling(p)
	return true
}

func doPollingTCP(p *pollingEnt){
	n,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	ok = false
	var rTime int64
	for i:=0  ; !ok && i <= p.Retry;i++{
		startTime := time.Now().UnixNano()
		conn, err := net.DialTimeout("tcp", n.IP +":" + p.Polling, time.Duration(p.Timeout) *time.Second)
		endTime := time.Now().UnixNano()
		if err != nil {
			astilog.Debugf("doPollingTCP err=%v",err)
			p.LastResult = fmt.Sprintf("%v",err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		p.LastResult = ""
		setPollingState(p,"normal")
	}	else {
		setPollingState(p,p.Level)
	}
	updatePolling(p)
}

func doPollingHTTP(p *pollingEnt){
	_,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	ok = false
	var rTime int64
	for i:=0  ; !ok && i <= p.Retry;i++{
		startTime := time.Now().UnixNano()
		err := doHTTPGet(p)
		endTime := time.Now().UnixNano()
		if err != nil {
			astilog.Debugf("doPollingHTTP err=%v",err)
			p.LastResult = fmt.Sprintf("%v",err)
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		setPollingState(p,"normal")
	}	else {
		setPollingState(p,p.Level)
	}
	updatePolling(p)
}

var insecureTransport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

var insecureClient = &http.Client{Transport: insecureTransport}

func doHTTPGet(p *pollingEnt) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Timeout) * time.Second)
	defer cancel()
	req, err := http.NewRequest(http.MethodGet,p.Polling, nil)
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

func doPollingTLS(p *pollingEnt){
	n,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	d := &net.Dialer{
		Timeout:time.Duration(p.Timeout) *time.Second,
	}
	ok = false
	var rTime int64
	var cs tls.ConnectionState
	for i:=0  ; !ok && i <= p.Retry;i++{
		startTime := time.Now().UnixNano()
		conn, err := tls.DialWithDialer(d,"tcp",n.IP +":"+ p.Polling, conf)
		endTime := time.Now().UnixNano()
		if err != nil {
			astilog.Debugf("doPollingTLS err=%v",err)
			p.LastResult = fmt.Sprintf("%v",err)
			continue
		}
		defer conn.Close()
		rTime = endTime - startTime
		cs = conn.ConnectionState()
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		p.LastResult = getTLSConnectioStateInfo(&cs)
		setPollingState(p,"normal")
	}	else {
		setPollingState(p,p.Level)
	}
	updatePolling(p)
}

var tlsCSMap = make(map[string]string)

func loadTLSParamsMap(path string) {
	file, err := os.Open(path)
	if err != nil {
		astilog.Errorf("loadTLSParamsMap err=%v",err)
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

func  getTLSConnectioStateInfo(cs *tls.ConnectionState) string{
	var v string
	switch cs.Version {
	case tls.VersionSSL30:
		v = "SSLv3"
	case tls.VersionTLS10:
		v = "TLSv1.0"
	case tls.VersionTLS11:
		v = "TLSv1.1"
	case tls.VersionTLS12:
		v = "TLSv1.2"
	case tls.VersionTLS13:
		v = "TLSv1.3"
	default:
		v = "Unknown"
	}
	id := fmt.Sprintf("%04x",cs.CipherSuite)
	if n,ok := tlsCSMap[id];ok {
		return fmt.Sprintf("%v %s",v,n)
	}
	return fmt.Sprintf("%v 0x%s",v,id)
}

func doPollingDNS(p *pollingEnt){
	_,ok := nodes[p.NodeID]
	if !ok {
		astilog.Errorf("node not found nodeID=%s",p.NodeID)
		return
	}
	ok = false
	var rTime int64
	var ip string
	for i:=0  ; !ok && i <= p.Retry;i++{
		startTime := time.Now().UnixNano()
		addr, err := net.ResolveIPAddr("ip", p.Polling)
		endTime := time.Now().UnixNano()
		if err != nil {
			astilog.Debugf("doPollingDNS err=%v",err)
			p.LastResult = fmt.Sprintf("ERR:%v",err)
			continue
		}
		rTime = endTime - startTime
		ok = true
		ip = addr.String()
	}
	if ok && p.LastResult != ""  && !strings.HasPrefix(p.LastResult,"ERR") && ip != p.LastResult {
		ok = false
	}
	p.LastVal = float64(rTime)
	if ok {
		p.LastResult = ip
		setPollingState(p,"normal")
	}	else {
		setPollingState(p,p.Level)
	}
	updatePolling(p)
}

