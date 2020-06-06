package main

import (
	"encoding/json"
	"strings"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
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
	case "showPolling":
		return showPolling(&m)
	case "deletePolling":
		return deletePollingMsg(&m)
	case "pollNow":
		return pollNow(&m)
	case "autoAddPolling":
		return autoAddPolling(&m)
	case "getTemplates":
		return pollingTemplates, nil
	}
	astiLogger.Errorf("Unknow Message Name=%s", m.Name)
	return "ok", nil
}

func getNodeBasicInfo(m *bootstrap.MessageIn) (interface{}, error) {
	var node nodeEnt
	if len(m.Payload) > 0 {
		var nodeID string
		if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
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
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
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
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
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
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if p.ID == "" {
			if err := addPolling(&p); err != nil {
				astiLogger.Errorf("addPolling %s error=%v", m.Name, err)
				return "ng", err
			}
		} else {
			p.LastResult = ""
			p.NextTime = 0
			p.State = "unkown"
			doPollingCh <- true
			if err := updatePolling(&p); err != nil {
				astiLogger.Errorf("updatePolling %s error=%v", m.Name, err)
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
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := deletePolling(id); err != nil {
			astiLogger.Errorf("deletePolling  error=%v", err)
			return "ng", err
		}
		if err := bootstrap.SendMessage(aiWindow, "deleteModel", id); err != nil {
			astiLogger.Errorf("sendSendMessage deleteModel error=%v", err)
		}
		return "ok", nil
	}
	return "ng", errInvalidNode
}

func pollNow(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if p, ok := pollings.Load(id); ok {
			p.(*pollingEnt).NextTime = 0
			p.(*pollingEnt).State = "unkown"
		} else {
			astiLogger.Errorf("No Polling")
			return "ng", nil
		}
		doPollingCh <- true
		return "ok", nil
	}
	return "ng", errInvalidNode
}

func showPolling(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		var ok bool
		var v interface{}
		params := struct {
			Polling *pollingEnt
			Node    *nodeEnt
		}{}
		if v, ok = pollings.Load(id); !ok {
			astiLogger.Errorf("No Polling id=%s", id)
			return "ng", nil
		}
		params.Polling = v.(*pollingEnt)
		if params.Node, ok = nodes[params.Polling.NodeID]; !ok {
			astiLogger.Errorf("No Node id=%s", params.Polling.NodeID)
			return "ng", nil
		}
		if err := bootstrap.SendMessage(pollingWindow, "setParams", params); err != nil {
			astiLogger.Errorf("sendSendMessage error=%v", err)
			return "ng", err
		}
		pollingWindow.Show()
		return "ok", nil
	}
	return "ng", errInvalidNode
}

func autoAddPolling(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		n, ok := nodes[id]
		if !ok {
			return "ng", nil
		}
		plist := getNodePollings(id)
		pmap := make(map[string]bool)
		for _, p := range plist {
			key := getSha1Key(p.Type + ":" + p.Polling)
			pmap[key] = true
		}
		for _, pt := range pollingTemplates {
			if pt.NodeType != "" {
				if n.Type == "" || !strings.Contains(n.Type, pt.NodeType) {
					continue
				}
			}
			key := getSha1Key(pt.Type + ":" + pt.Polling)
			if _, ok := pmap[key]; ok {
				continue
			}
			p := &pollingEnt{
				NodeID:  id,
				Name:    pt.Name,
				Type:    pt.Type,
				Level:   pt.Level,
				Polling: pt.Polling,
				State:   "unkown",
				PollInt: mapConf.PollInt,
				Timeout: mapConf.Timeout,
				Retry:   mapConf.Retry,
			}
			if err := addPolling(p); err != nil {
				astiLogger.Errorf("addPolling %s %v error=%v", m.Name, p, err)
				continue
			}
		}
		return "ok", nil
	}
	return "ng", errInvalidNode
}
