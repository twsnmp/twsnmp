package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	gosnmp "github.com/twsnmp/gosnmp"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	"github.com/sleepinggenius2/gosmi/parser"
)

// mibTreeEnt is
type mibTreeEnt struct {
	name     string
	oid      string
	children []*mibTreeEnt
}

var (
	mibTree     = map[string]*mibTreeEnt{}
	mibTreeRoot *mibTreeEnt
)

// mibMessageHandler handles messages
func mibMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		_ = mibWindow.Hide()
		return "ok", nil
	case "mibtree":
		return makeMibTreeJSON(), nil
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
		astiLogger.Errorf("snmpWalk err=%v", err)
		return ret, err
	}
	defer agent.Conn.Close()
	err = agent.Walk(mib.NameToOID(mibName), func(variable gosnmp.SnmpPDU) error {
		s := mib.OIDToName(variable.Name) + "="
		if variable.Type == gosnmp.OctetString {
			if strings.Contains(mib.OIDToName(variable.Name), "ifPhysAd") {
				a := variable.Value.(string)
				if len(a) > 5 {
					s += fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", a[0], a[1], a[2], a[3], a[4], a[5])
				}
			} else {
				s += variable.Value.(string)
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
		return fmt.Errorf("unmarshal %s error=%v", m.Name, err)
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
			return fmt.Errorf("can not find mib name %s", name)
		}
		a := strings.SplitN(oid, ".", 2)
		if len(a) < 2 {
			return fmt.Errorf("can not split mib name=%s oid=%s", name, oid)
		}
		noid, ok := mapNameToOID[a[0]]
		if !ok {
			return fmt.Errorf("can not split mib name=%s oid=%s", name, a[0])
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
		return fmt.Errorf("unmarshal %s error=%v", m.Name, err)
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
			_ = mib.Add(name, mapNameToOID[name])
		}
	}
	return nil
}

func addToMibTree(oid, name, poid string) {
	n := &mibTreeEnt{name: name, oid: oid, children: []*mibTreeEnt{}}
	if poid == "" {
		mibTreeRoot = n
	} else {
		p, ok := mibTree[poid]
		if !ok {
			astiLogger.Errorf("addToMibTree parentId=%v: not found", poid)
			return
		}
		p.children = append(p.children, n)
	}
	mibTree[oid] = n
}

func mibTreeJSONEnt(n *mibTreeEnt, prefix string) []string {
	r := []string{}
	r = append(r, fmt.Sprintf("%s\"name\":\"%v\",\"value\":\"%s\"", prefix, n.name, n.oid))
	if len(n.children) < 1 {
		return r
	}
	r = append(r, fmt.Sprintf("%s,\"children\": [", prefix))
	for i, c := range n.children {
		if i > 0 {
			r = append(r, fmt.Sprintf("%s,", prefix))
		}
		r = append(r, fmt.Sprintf("%s{", prefix))
		r = append(r, mibTreeJSONEnt(c, prefix+" ")...)
		r = append(r, fmt.Sprintf("%s}", prefix))
	}
	r = append(r, fmt.Sprintf("%s]", prefix))
	return r
}

func makeMibTreeJSON() string {
	oids := []string{}
	for _, n := range mib.GetNameList() {
		oid := mib.NameToOID(n)
		if oid == ".0.0" {
			continue
		}
		oids = append(oids, oid)
	}
	sort.Slice(oids, func(i, j int) bool {
		a := strings.Split(oids[i], ".")
		b := strings.Split(oids[j], ".")
		for k := 0; k < len(a) && k < len(b); k++ {
			l, _ := strconv.Atoi(a[k])
			m, _ := strconv.Atoi(b[k])
			if l == m {
				continue
			}
			if l < m {
				return true
			}
			return false
		}
		return len(a) < len(b)
	})
	addToMibTree(".1.3.6.1", "iso.org.dod.internet", "")
	for _, oid := range oids {
		name := mib.OIDToName(oid)
		if name == "" {
			continue
		}
		lastDot := strings.LastIndex(oid, ".")
		if lastDot < 0 {
			continue
		}
		poid := oid[:lastDot]
		addToMibTree(oid, name, poid)
	}
	if mibTreeRoot == nil {
		fmt.Printf("show: mibTreeRoot node not found\n")
		return ""
	}
	return "{\n" + strings.Join(mibTreeJSONEnt(mibTreeRoot, ""), "\n") + "}\n"
}
