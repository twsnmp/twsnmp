package main

import (
	"fmt"
	"encoding/json"
	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)


// nodeMessageHandler handles messages
func nodeMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "close":
			nodeWindow.Hide()
			return "ok",nil
		case "getNodeBasicInfo":
			var node nodeEnt
			if len(m.Payload) > 0 {
				var nodeID string
				if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if node,ok := nodes[nodeID]; ok {
					return node,nil
				}
			}
			return node, errInvalidNode
		case "getNodeLog":
			if len(m.Payload) > 0 {
				var nodeID string
				if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				return getNodeEventLogList(nodeID),nil
			}
			return "ng", errInvalidNode
		case "getNodePollings":
			if len(m.Payload) > 0 {
				var nodeID string
				if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				return getNodePollingList(nodeID),nil
			}
			return "ng", errInvalidNode
		case "savePolling":
			if len(m.Payload) > 0 {
				var p pollingEnt
				if err := json.Unmarshal(m.Payload, &p); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if p.ID == "" {
					if err := addPolling(&p); err != nil {
						astilog.Error(fmt.Sprintf("addPolling %s error=%v", m.Name, err))
						return "ng", err
					}
				} else {
					if err := updatePolling(&p); err != nil {
						astilog.Error(fmt.Sprintf("updatePolling %s error=%v", m.Name, err))
						return "ng", err
					}
				}
			}
			return "ok", nil
		case "deletePolling":
			if len(m.Payload) > 0 {
				var id string
				if err := json.Unmarshal(m.Payload, &id); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if err := deletePolling(id);err != nil {
					astilog.Error(fmt.Sprintf("deletePolling  error=%v", err))
					return "ng", err
				}
				return "ok",nil
			}
			return "ng", errInvalidNode
		case "pollNow":
			if len(m.Payload) > 0 {
				var id string
				if err := json.Unmarshal(m.Payload, &id); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if p,ok :=  pollings[id]; ok {
					p.LastTime = 0
					p.State = "unkown"
				} else {
					astilog.Error(fmt.Sprintf("No Polling"))
					return "ng", nil
				}
				return "ok",nil
			}
			return "ng", errInvalidNode
		}
	return "ok",nil
}

func getNodePollingList(nodeID string) []pollingEnt{
	ret := []pollingEnt{}
	for _,p := range pollings {
		if p.NodeID == nodeID {
			ret = append(ret,*p)
		} 
	}
	return ret
}