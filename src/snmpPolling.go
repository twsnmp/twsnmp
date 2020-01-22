package main

// snmpPolling.go : SNMPのポーリングを行う。

import (
	"time"
	"strings"
	"fmt"
	"strconv"
	gosnmp "github.com/soniah/gosnmp"

	astilog "github.com/asticode/go-astilog"

)

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
	if p.LastResult == "" {
		p.LastResult = fmt.Sprintf("sysUpTime=%d",uptime)
		return
	}
	var lastUptime int64
	if _,err := fmt.Sscanf(p.LastResult,"sysUpTime=%d",&lastUptime);err != nil {
		p.LastResult = fmt.Sprintf("sysUpTime=%d",uptime)
		p.LastVal = 0;
		setPollingState(p,"unkown")
	} else {
		p.LastVal = float64(uptime - lastUptime);
		p.LastResult = fmt.Sprintf("sysUpTime=%d",uptime)
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
		sv = fmt.Sprintf("%s=%d;sysUpTime=%d",m,iv,sut)
	}
	if mode == "ps" || mode == "delta" {
		if !strings.Contains(p.LastResult,";") {
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
			p.LastResult = fmt.Sprintf("%s=%d;sysUpTime=%d",m,iv,sut)
			setPollingState(p,"unkown")
			return
		}
		var liv int64
		var lsut int64
		var n1,n2 string
		n,err :=  fmt.Sscanf(p.LastResult,"%s=%d,%s=%d",&n1,&liv,&n2,lsut)
		if err != nil || n != 4 {
			p.LastResult = fmt.Sprintf("%s=%d;sysUpTime=%d",m,iv,sut)
			setPollingState(p,"unkown")
			return
		}
		if mode == "ps" {
			dsut := sut -  lsut
			if dsut <= 0 {
				p.LastResult = fmt.Sprintf("%s=%d;sysUpTime=%d",m,iv,sut)
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
