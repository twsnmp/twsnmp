package main

import (
	"fmt"
	"strings"
	"encoding/json"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)


// reportMessageHandler handles messages
func reportMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "close":
			reportWindow.Hide()
			return "ok",nil
		case "getDevices":
			return getDevices(),nil
		case "getUsers":
			return getUsers(),nil
		case "getServers":
			return getServers(),nil
		case "getFlows":
			return getFlows(),nil
		case "getAllow":
			return getAllow(),nil
		case "getDenny":
			return getDenny(),nil
		case "addRule":
			return addRuleEnt(&m)
		case "addRuleByID":
			return addRuleByIDEnt(&m)
		case "deleteAllow":
			return deleteAllow(&m)
		case "deleteDenny":
			return deleteDenny(&m)
		case "deleteDevice":
			return deleteReportEnt(&m,"devices")
		case "deleteUser":
			return deleteReportEnt(&m,"users")
		case "deleteServer":
			return deleteReportEnt(&m,"servers")
		case "deleteFlow":
			return deleteReportEnt(&m,"flows")
		case "resetReport":
			return resetReportEnt(&m)
	}
	astilog.Errorf("Unknow Message Name=%s",m.Name)
	return "ok",nil
}

func getDevices() []*deviceEnt {
	r := []*deviceEnt{}
	for _,d := range devices {
		r = append(r,d)
	} 
	return r
}

func getUsers() []*userEnt {
	r := []*userEnt{}
	for _,u := range users {
		r = append(r,u)
	} 
	return r
}

func getServers() []*serverEnt {
	r := []*serverEnt{}
	for _,s := range servers {
		r = append(r,s)
	} 
	return r
}

func getFlows() []*flowEnt {
	fl := []*flowEnt{}
	for _,d := range flows {
		fl = append(fl,d)
	}
	astilog.Debugf("getFlows len=%d", len(fl))
	return fl
}

type ruleEnt struct  {
	ID string
	Score  int
	Server string
	ServerName string
	Service string
	Loc     string
}

type ruleResEnt struct {
	Recomends []ruleEnt
	Rules     []ruleEnt
}

func getAllow() ruleResEnt {
	r := ruleResEnt{}
	for _,ar := range allowRules {
		for sv := range ar.Servers {
			r.Rules = append(r.Rules,ruleEnt{
				ID: fmt.Sprintf("%s:%s",sv,ar.Service),
				Server: sv,
				Service:ar.Service,
				ServerName: findNameFromIP(sv),
			})
		}
	}
	return r
}

func getDenny() ruleResEnt {
	r := ruleResEnt{}
	for id := range dennyRules {
		a := strings.Split(id,":")
		if len(a) == 3{
			r.Rules = append(r.Rules,ruleEnt{
				ID: id,
				Server: a[0],
				Service: a[1],
				ServerName: findNameFromIP(a[0]),
				Loc: a[2],
			})
		}
	}
	return r
}

func deleteReportEnt(m *bootstrap.MessageIn,r string) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := deleteReport(r,id); err != nil {
			astilog.Errorf("deleteDeviceFromReport  error=%v", err)
			return "ng", err
		}
	}
	return "ok", nil
}

func resetReportEnt(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var r string
		if err := json.Unmarshal(m.Payload, &r); err != nil {
			astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		resetPenalty(r)
	}
	return "ok", nil
}

type addRuleReq struct {
	Type string
	Server string
	Service string
	Loc string
}

func addRuleEnt(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var req addRuleReq
		if err := json.Unmarshal(m.Payload, &req); err != nil {
			astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		switch req.Type{
		case "allow_service":
			if err := addAllowRule(req.Service,req.Server);err != nil {
				return "ng",err
			}
		case "denny_service":
			if err := addDennyRule(fmt.Sprintf("*:%s:*",req.Service));err != nil {
				return "ng",err
			}
		case "denny_server":
			if err := addDennyRule(fmt.Sprintf("%s:*:*",req.Server));err != nil {
				return "ng",err
			}
		case "denny_server_service":
			if err := addDennyRule(fmt.Sprintf("%s:%s:*",req.Server,req.Service));err != nil {
				return "ng",err
			}
		case "denny_service_loc":
			if err := addDennyRule(fmt.Sprintf("*:%s:%s",req.Service,req.Loc));err != nil {
				return "ng",err
			}
		case "denny_loc":
			if err := addDennyRule(fmt.Sprintf("*:*:%s",req.Loc));err != nil {
				return "ng",err
			}
		}
	}
	return "ok", nil
}

func addRuleByIDEnt(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
	}
	return "ok", nil
}

func deleteAllow(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := deleteAllowRule(id);err != nil {
			astilog.Errorf("deleteAllowRule %s error=%v", m.Name, err)
			return "ng",err
		}
	}
	return "ok", nil
}

func deleteDenny(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := deleteDennyRule(id);err != nil {
			astilog.Errorf("deleteDennyRule %s error=%v", m.Name, err)
			return "ng",err
		}
	}
	return "ok", nil
}
