package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)

// mainWindowMessageHandler handles messages
func mainWindowMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "mapConf":
		{
			if len(m.Payload) > 0 {
				if err := json.Unmarshal(m.Payload, &mapConf); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if err := saveMapConfToDB(); err != nil {
					astilog.Error(fmt.Sprintf("saveMapConfToDB  error=%v", err))
					return "ng", err
				}
			}
			return "ok", nil
		}
	case "startDiscover":
		{
			if discoverStat.Running {
				return "ng", nil
			}
			if len(m.Payload) > 0 {
				if err := json.Unmarshal(m.Payload, &discoverConf); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if err := saveDiscoverConfToDB(); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				startDiscover()
			}
			return "ok", nil
		}
	case "getDiscover":
		return struct {
			Conf discoverConfEnt
			Stat discoverStatEnt
		} {
			Conf: discoverConf,
			Stat: discoverStat,
		}, nil
	case "stopDiscover":
		{
			go stopDiscover()
			return "ok", nil
		}
	case "saveNode":
		{
			if len(m.Payload) > 0 {
				var n nodeEnt
				if err := json.Unmarshal(m.Payload, &n); err != nil {
					astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
					return "ng", err
				}
				if n.ID == "" {
					if err := addNode(&n); err != nil {
						astilog.Errorf("addNode %s error=%v", m.Name, err)
						return "ng", err
					}
				} else {
					ntmp, ok := nodes[n.ID]
					if !ok {
						astilog.Errorf("%s invalid nodeid %s", m.Name, n.ID)
						return "ng", errInvalidNode
					}
					ntmp.Community = n.Community
					ntmp.Descr = n.Descr
					ntmp.IP = n.IP
					ntmp.Icon = n.Icon
					ntmp.Name = n.Name
					if err := updateNode(ntmp); err != nil {
						astilog.Errorf("editNode %s error=%v", m.Name, err)
						return "ng", err
					}
				}
			}
			applyMapData()
			return "ok", nil
		}
	case "updateNode":
		{
			var n nodeEnt
			if len(m.Payload) < 1 {
				return "ng", errNoPayload
			}
			if err := json.Unmarshal(m.Payload, &n); err != nil {
				astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return "ng", err
			}
			ntmp, ok := nodes[n.ID]
			if !ok {
				astilog.Error(fmt.Sprintf("%s  invalid node %s", m.Name, n.ID))
				return "ng", errInvalidNode
			}
			ntmp.X = n.X
			ntmp.Y = n.Y
			if err := updateNode(ntmp); err != nil {
				astilog.Error(fmt.Sprintf("editNode %s error=%v", m.Name, err))
				return "ng", err
			}
			return "ok", nil
		}
	case "deleteNode":
		{
			var nodeID string
			if len(m.Payload) < 1 {
				return "ng", errNoPayload
			}
			if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
				astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return "ng", err
			}
			if _, ok := nodes[nodeID]; !ok {
				return "ng", errInvalidNode
			}
			if err := deleteNode(nodeID); err != nil {
				return "ng", err
			}
			return "ok", nil
		}
	case "dupNode":
		{
			var nodeID string
			if len(m.Payload) < 1 {
				return "ng", errNoPayload
			}
			if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
				astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return "ng", err
			}
			if _, ok := nodes[nodeID]; !ok {
				return "ng", errInvalidNode
			}
			var n nodeEnt
			n.Name = nodes[nodeID].Name + "-Copy"
			n.IP = nodes[nodeID].IP
			n.Descr = nodes[nodeID].Descr
			n.Community = nodes[nodeID].Community
			n.X = nodes[nodeID].X + 32
			n.Y = nodes[nodeID].Y
			n.Icon = nodes[nodeID].Icon
			n.State = nodes[nodeID].State
			if err := addNode(&n); err != nil {
				astilog.Error(fmt.Sprintf("addNode %s error=%v", m.Name, err))
				return "ng", err
			}
			return n, nil
		}
	case "saveLine":
		{
			if len(m.Payload) > 0 {
				var l lineEnt
				if err := json.Unmarshal(m.Payload, &l); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if l.ID == "" {
					if err := addLine(&l); err != nil {
						astilog.Error(fmt.Sprintf("addLine %s error=%v", m.Name, err))
						return "ng", err
					}
				} else {
					if err := updateLine(&l); err != nil {
						astilog.Error(fmt.Sprintf("updateLine %s error=%v", m.Name, err))
						return "ng", err
					}
				}
				updateLineState()
			}
			applyMapData()
			return "ok", nil
		}
	case "deleteLine":
		{
			if len(m.Payload) > 0 {
				var l lineEnt
				if err := json.Unmarshal(m.Payload, &l); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if l.ID == "" {
					astilog.Error(fmt.Sprintf("delLine %s ", m.Name))
					return "ng", errInvalidID
				}
				if err := deleteLine(l.ID); err != nil {
					astilog.Error(fmt.Sprintf("deleteLine %s error=%v", m.Name, err))
					return "ng", err
				}
			}
			go applyMapData()
			return "ok", nil
		}
	case "getLine":
		{
			var l lineEnt
			if len(m.Payload) < 1 {
				return "ng", errNoPayload
			}
			if err := json.Unmarshal(m.Payload, &l); err != nil {
				astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return "ng", err
			}
			n1, ok := nodes[l.NodeID1]
			if !ok {
				astilog.Error(fmt.Sprintf("Invalid Node %s", m.Name))
				return "ng", errInvalidNode
			}
			n2, ok := nodes[l.NodeID2]
			if !ok {
				astilog.Error(fmt.Sprintf("Invalid Node %s", m.Name))
				return "ng", errInvalidNode
			}
			var dlgParam struct {
				Line      lineEnt
				NodeName1 string
				NodeName2 string
				Pollings1 []pollingEnt
				Pollings2 []pollingEnt
			}
			dlgParam.Line = l
			for _, ll := range lines {
				if ll.NodeID1 == l.NodeID1 && ll.NodeID2 == l.NodeID2 {
					dlgParam.Line = *ll
					break
				}
				if ll.NodeID1 == l.NodeID2 && ll.NodeID2 == l.NodeID1 {
					dlgParam.Line = *ll
					break
				}
			}
			dlgParam.NodeName1 = n1.Name
			dlgParam.NodeName2 = n2.Name
			dlgParam.Pollings1 = []pollingEnt{}
			dlgParam.Pollings2 = []pollingEnt{}
			for _, p := range pollings {
				if p.NodeID == l.NodeID1 {
					dlgParam.Pollings1 = append(dlgParam.Pollings1, *p)
				}
				if p.NodeID == l.NodeID2 {
					dlgParam.Pollings2 = append(dlgParam.Pollings2, *p)
				}
			}
			return dlgParam, nil
		}
	case"showNodeInfo":
		{
			var nodeID string
			if len(m.Payload) < 1 {
				return "ng", errNoPayload
			}
			if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
				astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
				return "ng", err
			}
			if err := bootstrap.SendMessage(nodeWindow, "setNodeID",nodeID); err != nil {
				astilog.Error(fmt.Sprintf("sendSendMessage %s error=%v",m.Name, err))
				return "ng",err
			}	
			nodeWindow.Show()
			return "ok", nil
		}
	case"logDisp":
		{
			logWindow.Show()
			logWindow.OpenDevTools()
			return "ok", nil
		}
	}
	return "ok", nil
}

func applyMapConf() {
	if err := bootstrap.SendMessage(mainWindow, "mapConf", mapConf); err != nil {
		astilog.Error(fmt.Sprintf("sendSendMessage mapConf error=%v", err))
		return
	}
}

func applyMapData() {
	if err := bootstrap.SendMessage(mainWindow, "nodes", nodes); err != nil {
		astilog.Error(fmt.Sprintf("sendSendMessage nodes error=%v", err))
		return
	}
	if err := bootstrap.SendMessage(mainWindow, "lines", lines); err != nil {
		astilog.Error(fmt.Sprintf("sendSendMessage lines error=%v", err))
		return
	}
}

func mainWindowBackend(ctx context.Context) {
	stateCheckNodes := make(map[string]bool)
	lastLog := sendLogs("")
	for k := range nodes {
		updateNodeState(k)
	}
	updateLineState()
	applyMapConf()
	applyMapData()
	for {
		select {
		case <-ctx.Done():
			return
		case p := <-pollingStateChangeCh:
			stateCheckNodes[p.NodeID] = true
		case <-time.Tick(time.Second * 10):
			lastLog = sendLogs(lastLog)
			if len(stateCheckNodes) > 0 {
				for k := range stateCheckNodes {
					updateNodeState(k)
					delete(stateCheckNodes, k)
				}
				updateLineState()
				applyMapData()
			}
		}
	}
}

func sendLogs(lastLog string) string {
	list := getEventLogList(lastLog, mapConf.LogDispSize)
	if len(list) > 0 {
		if err := bootstrap.SendMessage(mainWindow, "logs", list); err != nil {
			astilog.Error(fmt.Sprintf("sendSendMessage logs error=%v", err))
		} else {
			return fmt.Sprintf("%016x", list[0].Time)
		}
	}
	return lastLog
}

func updateNodeState(nodeID string) {
	n, ok := nodes[nodeID]
	if !ok {
		return
	}
	n.State = "unkown"
	for _, p := range pollings {
		if p.NodeID != nodeID {
			continue
		}
		if p.State == "high" {
			n.State = "high"
			break
		}
		if p.State == "low" {
			n.State = "low"
			continue
		}
		if n.State == "low" {
			continue
		}
		if p.State == "repair" {
			n.State = "repair"
		}
		if n.State == "repair" || n.State != "unkown" {
			continue
		}
		n.State = p.State
	}
}

func updateLineState() {
	for _, l := range lines {
		if p, ok := pollings[l.PollingID1]; ok {
			l.State1 = p.State
		}
		if p, ok := pollings[l.PollingID2]; ok {
			l.State2 = p.State
		}
	}
}
