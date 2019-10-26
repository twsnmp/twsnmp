package main

import (
	"encoding/json"
	"time"
	"fmt"
	"strings"

	gosnmp "github.com/soniah/gosnmp"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)

// mibMessageHandler handles messages
func mibMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "close":
			mibWindow.Hide()
			return "ok",nil
		case "get":
			return getMIB(&m)
		}
	astilog.Errorf("Unknow Message Name=%s",m.Name)
	return "ok",nil
}

func getMIB(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var param = struct {
			NodeID string
			Name   string
		}{}
		if err := json.Unmarshal(m.Payload, &param); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return []string{}, errInvalidParams
		}
		r,err := snmpWalk(param.NodeID,param.Name)
		if err != nil {
			return fmt.Sprintf("MIB取得できません。err=%v",err),err
		} 
		return r,nil
	}
	return []string{}, errInvalidParams
}

func snmpWalk(nodeID,mibName string) ([]string,error) {
	ret := []string{}
	n,ok := nodes[nodeID]
	if !ok {
		astilog.Errorf("snmpWalk Invalid nodeID %s ",nodeID)
		return ret,fmt.Errorf("snmpWalk Invalid nodeID %s ",nodeID)
	}
	agent := &gosnmp.GoSNMP{
		Target:             n.IP,
		Port:               161,
		Transport:          "udp",
		Community:          n.Community,
		Version:            gosnmp.Version2c,
		Timeout:            time.Duration(mapConf.Timeout) * time.Second,
		Retries:            mapConf.Retry,
		ExponentialTimeout: true,
		MaxOids:            gosnmp.MaxOids,
	}
	err := agent.Connect()
	if err != nil {
		astilog.Errorf("snmpWalk err=%v",err)
		return ret,err
	}
	defer agent.Conn.Close()
	err = agent.Walk(mib.NameToOID(mibName), func(variable gosnmp.SnmpPDU) error {
		s := mib.OIDToName(variable.Name) + "="
		if variable.Type == gosnmp.OctetString {
			if strings.Contains(mib.OIDToName(variable.Name),"ifPhysAd") {
				a := variable.Value.([]byte)
				if len(a) > 5{
					s += fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X",a[0],a[1],a[2],a[3],a[4],a[5])
				}
			} else {
				s += string(variable.Value.([]byte))
			}
		} else if variable.Type == gosnmp.ObjectIdentifier {
			s += mib.OIDToName(variable.Value.(string))
		} else {
			s += fmt.Sprintf("%d",gosnmp.ToBigInt(variable.Value).Int64())
		}
		ret = append(ret,s)
		return nil
	})
	return ret,err
}