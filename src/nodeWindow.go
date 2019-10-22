package main

import (
	"encoding/json"
	"fmt"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)

// nodeMessageHandler handles messages
func nodeMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		nodeWindow.Hide()
		return "ok", nil
	case "getNodeBasicInfo":
		return getNodeBasicInfo(&m)
	case "getNodeLog":
		return getNodeLog(&m)
	case "getNodePollings":
		return getNodePollingsMsg(&m)
	case "savePolling":
		return savePolling(&m)
	case "deletePolling":
		return deletePollingMsg(&m)
	case "pollNow":
		return pollNow(&m)
	}
	astilog.Errorf("Unknow Message Name=%s",m.Name)
	return "ok", nil
}

func getNodeBasicInfo(m *bootstrap.MessageIn) (interface{}, error) {
	var node nodeEnt
	if len(m.Payload) > 0 {
		var nodeID string
		if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return "ng", err
		}
		if node, ok := nodes[nodeID]; ok {
			return node, nil
		}
	}
	return node, errInvalidNode
}

func getNodeLog(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var nodeID string
		if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return "ng", err
		}
		return getNodeEventLogs(nodeID), nil
	}
	return "ng", errInvalidNode
}

func getNodePollingsMsg(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var nodeID string
		if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return "ng", err
		}
		return getNodePollings(nodeID), nil
	}
	return "ng", errInvalidNode
}

func savePolling(m *bootstrap.MessageIn) (interface{}, error) {
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
			p.LastResult = ""
			p.LastTime = 0
			p.State ="unkown"	
			if err := updatePolling(&p); err != nil {
				astilog.Error(fmt.Sprintf("updatePolling %s error=%v", m.Name, err))
				return "ng", err
			}
		}
	}
	return "ok", nil
}

func deletePollingMsg(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return "ng", err
		}
		if err := deletePolling(id); err != nil {
			astilog.Error(fmt.Sprintf("deletePolling  error=%v", err))
			return "ng", err
		}
		return "ok", nil
	}
	return "ng", errInvalidNode
}

func pollNow(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return "ng", err
		}
		if p, ok := pollings[id]; ok {
			p.LastTime = 0
			p.State = "unkown"
		} else {
			astilog.Error(fmt.Sprintf("No Polling"))
			return "ng", nil
		}
		return "ok", nil
	}
	return "ng", errInvalidNode
}
