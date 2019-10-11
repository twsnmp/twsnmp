package main

import (
	"fmt"
	"time"
	"encoding/json"
	"context"
	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)

var (
	errNoPayload = fmt.Errorf("No Payload")
	errInvalidNode = fmt.Errorf("Invalid Node")
)

// mainWindowMessageHandler handles messages
func mainWindowMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "configMap": 
			{
				if err := bootstrap.SendMessage(dialogWindow, "configMap",mapConf); err != nil {
					astilog.Error(fmt.Sprintf("sendSendMessage %s error=%v",m.Name, err))
					return "ng",err
				}	
				dialogWindow.Show()
				return "ok",nil
			}
		case "startDiscover": 
			{
				if discoverStat.Running {
					if err := bootstrap.SendMessage(dialogWindow, "discoverStat",discoverStat); err != nil {
						astilog.Error(fmt.Sprintf("sendSendMessage %s error=%v",m.Name, err))
						return "ng",err
					}	
				} else {
					if len(m.Payload) < 1 {
						return "ng",errNoPayload
					}
					if err := json.Unmarshal(m.Payload, &discoverStat); err != nil {
						astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
						return "ng",err
					}	
					if err := bootstrap.SendMessage(dialogWindow, "startDiscover",discoverConf); err != nil {
						astilog.Error(fmt.Sprintf("sendSendMessage %s error=%v",m.Name, err))
						return "ng",err
					}	
				}
				dialogWindow.Show()
				return "ok",nil
			}
		case "addNode":
			{
				var n nodeEnt
				if len(m.Payload) < 1 {
					return "ng",errNoPayload
				}
				if err := json.Unmarshal(m.Payload, &n); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return "ng",err
				}	
				n.Icon ="desktop"
				if err := bootstrap.SendMessage(dialogWindow, "editNode",n); err != nil {
					astilog.Error(fmt.Sprintf("SendMessage editNode error=%v",err))
					return "ng",err
				}	
				dialogWindow.Show()
				return "ok",nil
			}
		case "updateNode":
			{
				var n nodeEnt
				if len(m.Payload) < 1 {
					return "ng",errNoPayload
				}
				if err := json.Unmarshal(m.Payload, &n); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return "ng",err
				}	
				ntmp,ok  := nodes[n.ID]
				if !ok {
					astilog.Error(fmt.Sprintf("%s  invalid node %s",m.Name, n.ID))
					return "ng",errInvalidNode
				}
				ntmp.X = n.X
				ntmp.Y = n.Y
				if err := updateNode(ntmp);err != nil {
					astilog.Error(fmt.Sprintf("editNode %s error=%v",m.Name, err))
					return "ng",err
				}
				return "ok",nil
			}
		case "editNode":
			{
				var nodeID string 
				if len(m.Payload) < 1 {
					return "ng",errNoPayload
				}
				if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return "ng",err
				}
				n,ok := nodes[nodeID]
				if !ok {
					return "ng",errInvalidNode
				}
				if err := bootstrap.SendMessage(dialogWindow, "editNode",n); err != nil {
					astilog.Error(fmt.Sprintf("SendMessage editNode error=%v",err))
					return "ng",err
				}	
				dialogWindow.Show()
				return "ok",nil
			}
		case "deleteNode":
			{
				var nodeID string 
				if len(m.Payload) < 1 {
					return "ng",errNoPayload
				}
				if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return "ng",err
				}
				if _,ok := nodes[nodeID]; !ok {
					return "ng",errInvalidNode
				}
				if err := deleteNode(nodeID);err != nil {
					return "ng",err
				}
				return "ok",nil
			}
		case "dupNode":
			{
				var nodeID string 
				if len(m.Payload) < 1 {
					return "ng",errNoPayload
				}
				if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return "ng",err
				}
				if _,ok := nodes[nodeID]; !ok {
					return "ng",errInvalidNode
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
					astilog.Error(fmt.Sprintf("addNode %s error=%v",m.Name, err))
					return "ng",err
				}
				return n,nil
			}
		case "editLine":
			{
				var l lineEnt
				if len(m.Payload) < 1 {
					return "ng",errNoPayload
				}
				if err := json.Unmarshal(m.Payload, &l); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return "ng",err
				}
				n1,ok := nodes[l.NodeID1]
				if !ok {
					astilog.Error(fmt.Sprintf("Invalid Node %s",m.Name))
					return "ng",errInvalidNode
				}
				n2,ok := nodes[l.NodeID2]
				if !ok {
					astilog.Error(fmt.Sprintf("Invalid Node %s",m.Name))
					return "ng",errInvalidNode
				}
				var dlgParam struct {
					Line lineEnt
					NodeName1 string
					NodeName2 string
					Pollings1 []pollingEnt
					Pollings2 []pollingEnt
				}
				dlgParam.Line = l
				for _,ll := range lines {
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
				for _,p := range pollings {
					if p.NodeID == l.NodeID1 {
						dlgParam.Pollings1 = append(dlgParam.Pollings1,*p)
					}
					if p.NodeID == l.NodeID2 {
						dlgParam.Pollings2 = append(dlgParam.Pollings2,*p)
					}
				}
				if err := bootstrap.SendMessage(dialogWindow, "editLine",dlgParam); err != nil {
					astilog.Error(fmt.Sprintf("SendMessage editLine error=%v",err))
					return "ng",err
				}	
				dialogWindow.Show()
				dialogWindow.Center()
				return "ok",nil
			}
	}
	return "ok",nil
}

func applyMapConf() {
	if err := bootstrap.SendMessage(mainWindow, "mapConf",mapConf); err != nil {
		astilog.Error(fmt.Sprintf("sendSendMessage mapConf error=%v", err))
		return
	}	
}

func applyMapData() {
	if err := bootstrap.SendMessage(mainWindow, "nodes",nodes); err != nil {
		astilog.Error(fmt.Sprintf("sendSendMessage nodes error=%v", err))
		return
	}	
	if err := bootstrap.SendMessage(mainWindow, "lines",lines); err != nil {
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
		case <- ctx.Done():
			return
		case p := <- pollingStateChangeCh:
			stateCheckNodes[p.NodeID] = true
		case <- time.Tick(time.Second * 5):
			lastLog = sendLogs(lastLog)
			if len(stateCheckNodes) > 0  {
				for k := range stateCheckNodes{
					updateNodeState(k)
					delete(stateCheckNodes,k)
				}
				updateLineState()
				applyMapData()
			}
		}
	}
}

func sendLogs(lastLog string) string {
	list := getEventLogList(lastLog,mapConf.LogDispSize)
	if len(list) > 0 {
		if err := bootstrap.SendMessage(mainWindow,"logs",list); err != nil {
			astilog.Error(fmt.Sprintf("sendSendMessage logs error=%v",err))
		} else {
			return fmt.Sprintf("%016x",list[0].Time)
		}
	}
	return lastLog
}

func updateNodeState(nodeID string){
	n,ok := nodes[nodeID]
	if !ok {
		return
	}
	n.State = "unkown"
	for _,p := range pollings {
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
		if p.State == "repair"  {
			n.State = "repair"
		}
		if n.State == "repair" || n.State != "unkown" {
			continue
		}
		n.State = p.State
	}
}

func updateLineState() {
	for _,l := range lines {
		if p,ok := pollings[l.PollingID1]; ok {
			l.State1 = p.State
		}
		if p,ok := pollings[l.PollingID2]; ok {
			l.State2 = p.State
		}
	}
}