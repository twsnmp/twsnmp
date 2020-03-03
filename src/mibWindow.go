package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	gosnmp "github.com/soniah/gosnmp"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/sleepinggenius2/gosmi/parser"
)

// mibMessageHandler handles messages
func mibMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		mibWindow.Hide()
		return "ok", nil
	case "get":
		return getMIB(&m)
	}
	astiLogger.Errorf("Unknow Message Name=%s", m.Name)
	return "ok", nil
}

func getMIB(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var param = struct {
			NodeID string
			Name   string
		}{}
		if err := json.Unmarshal(m.Payload, &param); err != nil {
			astiLogger.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return []string{}, errInvalidParams
		}
		r, err := snmpWalk(param.NodeID, param.Name)
		if err != nil {
			return fmt.Sprintf("MIB取得できません。err=%v", err), err
		}
		return r, nil
	}
	return []string{}, errInvalidParams
}

func snmpWalk(nodeID, mibName string) ([]string, error) {
	ret := []string{}
	n, ok := nodes[nodeID]
	if !ok {
		astiLogger.Errorf("snmpWalk Invalid nodeID %s ", nodeID)
		return ret, fmt.Errorf("snmpWalk Invalid nodeID %s ", nodeID)
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
		astiLogger.Errorf("snmpWalk err=%v", err)
		return ret, err
	}
	defer agent.Conn.Close()
	err = agent.Walk(mib.NameToOID(mibName), func(variable gosnmp.SnmpPDU) error {
		s := mib.OIDToName(variable.Name) + "="
		if variable.Type == gosnmp.OctetString {
			if strings.Contains(mib.OIDToName(variable.Name), "ifPhysAd") {
				a := variable.Value.([]byte)
				if len(a) > 5 {
					s += fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", a[0], a[1], a[2], a[3], a[4], a[5])
				}
			} else {
				s += string(variable.Value.([]byte))
			}
		} else if variable.Type == gosnmp.ObjectIdentifier {
			s += mib.OIDToName(variable.Value.(string))
		} else {
			s += fmt.Sprintf("%d", gosnmp.ToBigInt(variable.Value).Int64())
		}
		ret = append(ret, s)
		return nil
	})
	return ret, err
}

func addMIBFile(m *bootstrap.MessageIn) error {
	if len(m.Payload) < 1 {
		return errInvalidParams
	}
	var path string
	if err := json.Unmarshal(m.Payload, &path); err != nil {
		return fmt.Errorf("Unmarshal %s error=%v", m.Name, err)
	}
	var nameList []string
	var mapNameToOID = make(map[string]string)
	for _, name := range mib.GetNameList() {
		mapNameToOID[name] = mib.NameToOID(name)
	}
	module, err := parser.ParseFile(path)
	if err != nil {
		return fmt.Errorf("ParseFile %s error=%v", m.Name, err)
	}
	key := module.Name.String()
	if module.Body.Identity != nil {
		name := module.Body.Identity.Name.String()
		oid := getOid(&module.Body.Identity.Oid)
		mapNameToOID[name] = oid
		nameList = append(nameList, name)
	}
	for _, n := range module.Body.Nodes {
		name := n.Name.String()
		mapNameToOID[name] = getOid(n.Oid)
		nameList = append(nameList, name)
	}
	for _, name := range nameList {
		oid, ok := mapNameToOID[name]
		if !ok {
			return fmt.Errorf("Can not find mib name %s", name)
		}
		a := strings.SplitN(oid, ".", 2)
		if len(a) < 2 {
			return fmt.Errorf("Can not split mib name=%s oid=%s", name, oid)
		}
		noid, ok := mapNameToOID[a[0]]
		if !ok {
			return fmt.Errorf("Can not split mib name=%s oid=%s", name, a[0])
		}
		mapNameToOID[name] = noid + "." + a[1]
	}
	return putMIBFileToDB(key, path)
}

func delMIBModule(m *bootstrap.MessageIn) error {
	if len(m.Payload) < 1 {
		return errInvalidParams
	}
	var key string
	if err := json.Unmarshal(m.Payload, &key); err != nil {
		return fmt.Errorf("Unmarshal %s error=%v", m.Name, err)
	}
	return delMIBModuleFromDB(key)
}

func getOid(oid *parser.Oid) string {
	ret := ""
	for _, o := range oid.SubIdentifiers {
		if o.Name != nil {
			ret += o.Name.String()
		}
		if o.Number != nil {
			ret += fmt.Sprintf(".%d", int(*o.Number))
		}
	}
	return ret
}

func loadMIBDB() error {
	var nameList []string
	var mapNameToOID = make(map[string]string)
	for _, name := range mib.GetNameList() {
		mapNameToOID[name] = mib.NameToOID(name)
	}
	for _, m := range getMIBModuleList() {
		asn1 := getMIBModule(m)
		module, err := parser.Parse(bytes.NewReader(asn1))
		if err != nil || module == nil {
			continue
		}
		if module.Body.Identity != nil {
			name := module.Body.Identity.Name.String()
			oid := getOid(&module.Body.Identity.Oid)
			mapNameToOID[name] = oid
			nameList = append(nameList, name)
		}
		for _, n := range module.Body.Nodes {
			name := n.Name.String()
			mapNameToOID[name] = getOid(n.Oid)
			nameList = append(nameList, name)
		}
		for _, name := range nameList {
			oid, ok := mapNameToOID[name]
			if !ok {
				astiLogger.Errorf("Can not find mib name %s", name)
				continue
			}
			a := strings.SplitN(oid, ".", 2)
			if len(a) < 2 {
				astiLogger.Errorf("Can not split mib name=%s oid=%s", name, oid)
				continue
			}
			noid, ok := mapNameToOID[a[0]]
			if !ok {
				astiLogger.Errorf("Can not split mib name=%s oid=%s", name, a[0])
				continue
			}
			mapNameToOID[name] = noid + "." + a[1]
		}
		for _, name := range nameList {
			mib.Add(name, mapNameToOID[name])
		}
	}
	return nil
}
