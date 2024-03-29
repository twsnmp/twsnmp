package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

// reportMessageHandler handles messages
func reportMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		_=reportWindow.Hide()
		return "ok", nil
	case "getDevices":
		return getDevices(), nil
	case "getUsers":
		return getUsers(), nil
	case "getServers":
		return getServers(), nil
	case "getFlows":
		return getFlows(), nil
	case "getRules":
		return getRules(), nil
	case "addRule":
		return addRuleEnt(&m)
	case "deleteRule":
		return deleteRuleEnt(&m)
	case "deleteDevice":
		return deleteReportEnt(&m, "devices")
	case "deleteUser":
		return deleteReportEnt(&m, "users")
	case "deleteServer":
		return deleteReportEnt(&m, "servers")
	case "deleteFlow":
		return deleteReportEnt(&m, "flows")
	case "resetReport":
		return resetReportEnt(&m)
	case "showLoc":
		return openURL(&m)
	case "getIPInfo":
		return doGetIPInfo(&m)
	case "inquiryAddr":
		return doInquiryAddr(&m)
	}
	astiLogger.Errorf("Unknow Message Name=%s", m.Name)
	return "ok", nil
}

func getDevices() []*deviceEnt {
	r := []*deviceEnt{}
	for _, d := range devices {
		r = append(r, d)
	}
	return r
}

func getUsers() []*userEnt {
	r := []*userEnt{}
	for _, u := range users {
		r = append(r, u)
	}
	return r
}

func getServers() []*serverEnt {
	r := []*serverEnt{}
	for _, s := range servers {
		r = append(r, s)
	}
	return r
}

func getFlows() []*flowEnt {
	fl := []*flowEnt{}
	for _, d := range flows {
		fl = append(fl, d)
	}
	astiLogger.Debugf("getFlows len=%d", len(fl))
	return fl
}

type ruleEnt struct {
	Type       string
	ID         string
	Server     string
	ServerName string
	Service    string
	Loc        string
}

func getRules() []ruleEnt {
	r := []ruleEnt{}
	for _, ar := range allowRules {
		for sv := range ar.Servers {
			r = append(r, ruleEnt{
				Type:       "allow",
				ID:         fmt.Sprintf("allow:%s:%s", sv, ar.Service),
				Server:     sv,
				Service:    ar.Service,
				ServerName: findNameFromIP(sv),
			})
		}
	}
	for id := range dennyRules {
		a := strings.Split(id, ":")
		if len(a) == 3 {
			r = append(r, ruleEnt{
				Type:       "denny",
				ID:         "denny:" + id,
				Server:     a[0],
				Service:    a[1],
				ServerName: findNameFromIP(a[0]),
				Loc:        a[2],
			})
		}
	}
	return r
}

func deleteReportEnt(m *bootstrap.MessageIn, r string) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := deleteReport(r, id); err != nil {
			astiLogger.Errorf("deleteDeviceFromReport  error=%v", err)
			return "ng", err
		}
	}
	return "ok", nil
}

func resetReportEnt(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var r string
		if err := json.Unmarshal(m.Payload, &r); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		resetPenalty(r)
	}
	return "ok", nil
}

type addRuleReq struct {
	Type    string
	Server  string
	Service string
	Loc     string
}

func addRuleEnt(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var req addRuleReq
		if err := json.Unmarshal(m.Payload, &req); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		switch req.Type {
		case "allow_service":
			if err := addAllowRule(req.Service, req.Server); err != nil {
				return "ng", err
			}
		case "denny_service":
			if err := addDennyRule(fmt.Sprintf("*:%s:*", req.Service)); err != nil {
				return "ng", err
			}
		case "denny_server":
			if err := addDennyRule(fmt.Sprintf("%s:*:*", req.Server)); err != nil {
				return "ng", err
			}
		case "denny_server_service":
			if err := addDennyRule(fmt.Sprintf("%s:%s:*", req.Server, req.Service)); err != nil {
				return "ng", err
			}
		case "denny_service_loc":
			if err := addDennyRule(fmt.Sprintf("*:%s:%s", req.Service, req.Loc)); err != nil {
				return "ng", err
			}
		case "denny_loc":
			if err := addDennyRule(fmt.Sprintf("*:*:%s", req.Loc)); err != nil {
				return "ng", err
			}
		}
	}
	return "ok", nil
}

func deleteRuleEnt(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		a := strings.SplitN(id, ":", 2)
		if len(a) != 2 {
			return "ng", nil
		}
		if strings.Contains(a[0], "allow") {
			if err := deleteAllowRule(id); err != nil {
				astiLogger.Errorf("deleteAllowRule %s error=%v", m.Name, err)
				return "ng", err
			}
		} else {
			if err := deleteDennyRule(id); err != nil {
				astiLogger.Errorf("deleteDennyRule %s error=%v", m.Name, err)
				return "ng", err
			}
		}
	}
	return "ok", nil
}

func doGetIPInfo(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) < 3 {
		astiLogger.Errorf("doGetIPInfo %s payload=%v", m.Name, m.Payload)
		return "ng", nil
	}
	var ip string
	if err := json.Unmarshal(m.Payload, &ip); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	return getIPInfo(ip), nil
}

type inquiryAddrEnt struct {
	Mode string
	Name string
	Addr string
}

func doInquiryAddr(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) < 3 {
		astiLogger.Errorf("doNewAddr %s payload=%v", m.Name, m.Payload)
		return "ng", nil
	}
	var na inquiryAddrEnt
	if err := json.Unmarshal(m.Payload, &na); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	//	astiLogger.Infof("%v", na)
	now := time.Now().UnixNano()
	if na.Mode == "servers" {
		if _, ok := servers[na.Addr]; ok {
			return "dup", nil
		}
		servers[na.Addr] = &serverEnt{
			ID:         na.Addr,
			Server:     na.Addr,
			Services:   make(map[string]int64),
			ServerName: fmt.Sprintf("%s(%s)", na.Name, findNameFromIP(na.Addr)),
			Loc:        getLoc(na.Addr),
			Count:      1,
			Bytes:      0,
			FirstTime:  now,
			LastTime:   now,
			UpdateTime: now,
		}
	} else if na.Mode == "devices" {
		mac := normMACAddr(na.Addr)
		_, ok := devices[mac]
		if ok {
			return "dup", nil
		}
		devices[mac] = &deviceEnt{
			ID:         mac,
			IP:         "",
			Name:       na.Name,
			Vendor:     oui.Find(mac),
			FirstTime:  now,
			LastTime:   now,
			UpdateTime: now,
		}
	}
	return "ok", nil
}
