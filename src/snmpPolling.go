package main

// snmpPolling.go : SNMPのポーリングを行う。

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	gosnmp "github.com/twsnmp/gosnmp"
)

func doPollingSnmp(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		astiLogger.Errorf("node not found nodeID=%s", p.NodeID)
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
	if n.SnmpMode != "" {
		agent.Version = gosnmp.Version3
		agent.SecurityModel = gosnmp.UserSecurityModel
		if n.SnmpMode == "v3auth" {
			agent.MsgFlags = gosnmp.AuthNoPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 n.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: n.Password,
			}
		} else {
			agent.MsgFlags = gosnmp.AuthPriv
			agent.SecurityParameters = &gosnmp.UsmSecurityParameters{
				UserName:                 n.User,
				AuthenticationProtocol:   gosnmp.SHA,
				AuthenticationPassphrase: n.Password,
				PrivacyProtocol:          gosnmp.AES,
				PrivacyPassphrase:        n.Password,
			}
		}
	}
	err := agent.Connect()
	if err != nil {
		astiLogger.Errorf("SNMP agent.Connect err=%v", err)
		return
	}
	defer agent.Conn.Close()
	mode, params := parseSnmpPolling(p.Polling)
	if mode == "" {
		setPollingError("snmp", p, fmt.Errorf("invalid snmp polling"))
		return
	}
	if mode == "sysUpTime" {
		doPollingSnmpSysUpTime(p, agent)
	} else if strings.HasPrefix(mode, "ifOperStatus.") {
		doPollingSnmpIF(p, mode, agent)
	} else if mode == "count" {
		doPollingSnmpCount(p, params, agent)
	} else if mode == "process" {
		doPollingSnmpProcess(p, params, agent)
	} else if mode == "stats" {
		doPollingSnmpStats(p, params, agent)
	} else {
		doPollingSnmpGet(p, mode, params, agent)
	}
}

func parseSnmpPolling(s string) (string, string) {
	a := strings.SplitN(s, "|", 2)
	if len(a) < 1 {
		return "", ""
	}
	if len(a) < 2 {
		return strings.TrimSpace(a[0]), ""
	}
	return strings.TrimSpace(a[0]), strings.TrimSpace(a[1])
}

func doPollingSnmpSysUpTime(p *pollingEnt, agent *gosnmp.GoSNMP) {
	oids := []string{mib.NameToOID("sysUpTime.0")}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmpUpTime", p, err)
		return
	}
	var uptime int64
	for _, variable := range result.Variables {
		if variable.Name == mib.NameToOID("sysUpTime.0") {
			uptime = int64(gosnmp.ToBigInt(variable.Value).Uint64())
			break
		}
	}
	if uptime == 0 {
		setPollingError("snmpUpTime", p, fmt.Errorf("uptime==0"))
		return
	}
	lr := make(map[string]string)
	_ = json.Unmarshal([]byte(p.LastResult), &lr)
	if lut, ok := lr["sysUpTime"]; ok {
		lastUptime, err := strconv.ParseInt(lut, 10, 64)
		if err != nil {
			delete(lr, "sysUpTime")
			p.LastResult = makeLastResult(lr)
			setPollingError("snmp", p, err)
			return
		}
		p.LastVal = float64(uptime - lastUptime)
		lr["sysUpTime"] = fmt.Sprintf("%d", uptime)
		p.LastResult = makeLastResult(lr)
		if lastUptime < uptime {
			setPollingState(p, "normal")
			return
		}
		setPollingState(p, p.Level)
		return
	}
	p.LastVal = 0.0
	lr["sysUpTime"] = fmt.Sprintf("%d", uptime)
	p.LastResult = makeLastResult(lr)
	setPollingState(p, "unknown")
}

func doPollingSnmpIF(p *pollingEnt, ps string, agent *gosnmp.GoSNMP) {
	a := strings.Split(ps, ".")
	if len(a) < 2 {
		setPollingError("snmpif", p, fmt.Errorf("invalid format"))
		return
	}
	oids := []string{mib.NameToOID("ifOperStatus." + a[1]), mib.NameToOID("ifAdminStatus." + a[1])}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmpif", p, err)
		return
	}
	var oper int64
	var admin int64
	for _, variable := range result.Variables {
		if strings.HasPrefix(mib.OIDToName(variable.Name), "ifOperStatus") {
			oper = gosnmp.ToBigInt(variable.Value).Int64()
		} else if strings.HasPrefix(mib.OIDToName(variable.Name), "ifAdminStatus") {
			admin = gosnmp.ToBigInt(variable.Value).Int64()
		}
	}
	lr := make(map[string]string)
	p.LastVal = float64(oper)
	lr["oper"] = fmt.Sprintf("%d", oper)
	lr["admin"] = fmt.Sprintf("%d", admin)
	p.LastResult = makeLastResult(lr)
	if oper == 1 {
		setPollingState(p, "normal")
		return
	} else if admin == 2 {
		setPollingState(p, "normal")
		return
	} else if oper == 2 && admin == 1 {
		setPollingState(p, p.Level)
		return
	}
	setPollingState(p, "unknown")
}

func doPollingSnmpGet(p *pollingEnt, mode, params string, agent *gosnmp.GoSNMP) {
	a := strings.Split(params, "|")
	if len(a) < 2 {
		setPollingError("snmp", p, fmt.Errorf("invalid format"))
		return
	}
	names := strings.Split(a[0], ",")
	script := a[1]
	oids := []string{}
	for _, n := range names {
		if n == "" {
			continue
		}
		if oid := mib.NameToOID(n); oid != "" {
			oids = append(oids, strings.TrimSpace(oid))
		}
	}
	if len(oids) < 1 {
		setPollingError("snmp", p, fmt.Errorf("invalid format"))
		return
	}
	if mode == "ps" {
		oids = append(oids, mib.NameToOID("sysUpTime.0"))
	}
	result, err := agent.Get(oids)
	if err != nil {
		setPollingError("snmp", p, err)
		return
	}
	vm := otto.New()
	lr := make(map[string]string)
	for _, variable := range result.Variables {
		if variable.Name == mib.NameToOID("sysUpTime.0") {
			sut := gosnmp.ToBigInt(variable.Value).Uint64()
			_ = vm.Set("sysUpTime", sut)
			lr["sysUpTime.0"] = fmt.Sprintf("%d", sut)
			if mode == "ps" || mode == "delta" {
				lr["sysUpTime.0_Last"] = fmt.Sprintf("%d", sut)
			}
			continue
		}
		n := mib.OIDToName(variable.Name)
		vn := getValueName(n)
		if variable.Type == gosnmp.OctetString {
			v := variable.Value.(string)
			_ = vm.Set(vn, v)
			lr[n] = v
		} else if variable.Type == gosnmp.ObjectIdentifier {
			v := mib.OIDToName(variable.Value.(string))
			_ = vm.Set(vn, v)
			lr[n] = v
		} else {
			v := gosnmp.ToBigInt(variable.Value).Uint64()
			_ = vm.Set(vn, v)
			lr[n] = fmt.Sprintf("%d", v)
			if mode == "ps" || mode == "delta" {
				lr[n+"_Last"] = lr[n]
			}
		}
	}
	if mode == "ps" || mode == "delta" {
		oldlr := make(map[string]string)
		if err := json.Unmarshal([]byte(p.LastResult), &oldlr); err != nil || oldlr["error"] != "" {
			p.LastResult = makeLastResult(lr)
			setPollingState(p, "unknown")
			return
		}
		nvmap := make(map[string]int64)
		for k, v := range lr {
			if strings.HasPrefix(k, "_Last") {
				continue
			}
			if vo, ok := oldlr[k+"_Last"]; ok {
				if nv, err := strconv.ParseInt(v, 10, 64); err == nil {
					if nvo, err := strconv.ParseInt(vo, 10, 64); err == nil {
						nvmap[k] = nv - nvo
					}
				}
			}
		}
		sut := float64(1.0)
		if mode == "ps" {
			v, ok := nvmap["sysUpTime.0"]
			if !ok || v == 0 {
				setPollingError("snmp", p, fmt.Errorf("invalid format %v", nvmap))
				return
			}
			sut = float64(v)
		}
		for k, v := range nvmap {
			lr[k] = fmt.Sprintf("%f", float64(v*100.0)/sut)
			vn := getValueName(k)
			_ = vm.Set(vn, float64(v*100.0)/sut)
		}
	}
	value, err := vm.Run(script)
	if err == nil {
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					p.LastVal = vf
				}
			}
		}
		p.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(p, p.Level)
			return
		}
		setPollingState(p, "normal")
		return
	}
	setPollingError("snmp", p, err)
}

func getValueName(n string) string {
	a := strings.SplitN(n, ".", 2)
	return (a[0])
}

func doPollingSnmpCount(p *pollingEnt, params string, agent *gosnmp.GoSNMP) {
	cmds := splitCmd(params)
	if len(cmds) < 3 {
		setPollingError("snmp", p, fmt.Errorf("invalid format"))
		return
	}
	oid := mib.NameToOID(cmds[0])
	filter := parseFilter(cmds[1])
	script := cmds[2]
	count := 0
	var regexFilter *regexp.Regexp
	var err error
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			astiLogger.Errorf("doPollingSnmpCount err=%v", err)
			regexFilter = nil
		}
	}
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		s := ""
		if variable.Type == gosnmp.OctetString {
			if strings.Contains(mib.OIDToName(variable.Name), "ifPhysAd") {
				a := variable.Value.(string)
				if len(a) > 5 {
					s = fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", a[0], a[1], a[2], a[3], a[4], a[5])
				}
			} else {
				s = variable.Value.(string)
			}
		} else if variable.Type == gosnmp.ObjectIdentifier {
			s = mib.OIDToName(variable.Value.(string))
		} else {
			s = fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Uint64())
		}
		if regexFilter != nil && !regexFilter.Match([]byte(s)) {
			return nil
		}
		count++
		return nil
	}); err != nil {
		setPollingError("snmp", p, err)
		return
	}
	vm := otto.New()
	lr := make(map[string]string)
	_ = vm.Set("count", count)
	lr["count"] = fmt.Sprintf("%d", count)
	value, err := vm.Run(script)
	if err == nil {
		p.LastVal = float64(count)
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					p.LastVal = vf
				}
			}
		}
		p.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(p, p.Level)
			return
		}
		setPollingState(p, "normal")
		return
	}
	setPollingError("snmp", p, err)
}

func doPollingSnmpProcess(p *pollingEnt, params string, agent *gosnmp.GoSNMP) {
	cmds := splitCmd(params)
	if len(cmds) < 2 {
		setPollingError("snmp", p, fmt.Errorf("doPollingSnmpProcess Invalid format"))
		return
	}
	oid := mib.NameToOID("hrSWRunName")
	filter := parseFilter(cmds[0])
	script := cmds[1]
	var regexFilter *regexp.Regexp
	var err error
	if filter != "" {
		if regexFilter, err = regexp.Compile(filter); err != nil {
			astiLogger.Errorf("doPollingSnmpProcess err=%v", err)
			regexFilter = nil
		}
	}
	lastPidSum := 0
	lr := make(map[string]string)
	if err := json.Unmarshal([]byte(p.LastResult), &lr); err == nil {
		if s, ok := lr["pidSum"]; ok {
			if n, err := strconv.Atoi(s); err == nil {
				lastPidSum = n
			}
		}
	}
	pidSum := 0
	count := 0
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		if variable.Type != gosnmp.OctetString {
			return nil
		}
		n := mib.OIDToName(variable.Name)
		a := strings.SplitN(n, ".", 2)
		s := variable.Value.(string)
		if len(a) != 2 || a[0] != "hrSWRunName" {
			return nil
		}
		pid, err := strconv.Atoi(a[1])
		if err != nil {
			return nil
		}
		if regexFilter != nil && !regexFilter.Match([]byte(s)) {
			return nil
		}
		pidSum += pid
		count++
		return nil
	}); err != nil {
		setPollingError("snmp", p, err)
		return
	}
	changed := 0
	if lastPidSum != 0 && pidSum != lastPidSum {
		changed = 1
	}
	vm := otto.New()
	_ = vm.Set("count", count)
	_ = vm.Set("changed", changed)
	lr["count"] = fmt.Sprintf("%d", count)
	lr["pidSum"] = fmt.Sprintf("%d", pidSum)
	lr["changed"] = fmt.Sprintf("%d", changed)
	value, err := vm.Run(script)
	if err == nil {
		p.LastVal = float64(count)
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					p.LastVal = vf
				}
			}
		}
		p.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(p, p.Level)
			return
		}
		setPollingState(p, "normal")
		return
	}
	setPollingError("snmp", p, err)
}

func doPollingSnmpStats(p *pollingEnt, params string, agent *gosnmp.GoSNMP) {
	cmds := splitCmd(params)
	if len(cmds) < 2 {
		setPollingError("snmp", p, fmt.Errorf("invalid format"))
		return
	}
	oid := mib.NameToOID(cmds[0])
	script := cmds[1]
	count := uint64(0)
	sum := uint64(0)
	if err := agent.Walk(oid, func(variable gosnmp.SnmpPDU) error {
		if variable.Type != gosnmp.Counter32 &&
			variable.Type != gosnmp.Counter64 &&
			variable.Type != gosnmp.Integer &&
			variable.Type != gosnmp.Uinteger32 &&
			variable.Type != gosnmp.Gauge32 {
			return fmt.Errorf("mib is not number %#v", variable)
		}
		sum += gosnmp.ToBigInt(variable.Value).Uint64()
		count++
		return nil
	}); err != nil {
		setPollingError("snmp", p, err)
		return
	}
	if count < 1 {
		setPollingError("snmp", p, fmt.Errorf("no data"))
		return
	}
	avg := float64(sum) / float64(count)
	vm := otto.New()
	lr := make(map[string]string)
	_ = vm.Set("count", count)
	_ = vm.Set("sum", sum)
	_ = vm.Set("avg", avg)
	lr["count"] = fmt.Sprintf("%d", count)
	lr["sum"] = fmt.Sprintf("%d", sum)
	lr["avg"] = fmt.Sprintf("%f", avg)
	value, err := vm.Run(script)
	if err == nil {
		p.LastVal = float64(avg)
		if v, err := vm.Get("numVal"); err == nil {
			if v.IsNumber() {
				if vf, err := v.ToFloat(); err == nil {
					p.LastVal = vf
				}
			}
		}
		p.LastResult = makeLastResult(lr)
		if ok, _ := value.ToBoolean(); !ok {
			setPollingState(p, p.Level)
			return
		}
		setPollingState(p, "normal")
		return
	}
	setPollingError("snmp", p, err)
}
