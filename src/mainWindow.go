package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

// mainWindowMessageHandler handles messages
func mainWindowMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "getMIBModuleList":
		return getMIBModuleList(), nil
	case "addMIBFile":
		if err := addMIBFile(&m); err != nil {
			astiLogger.Error(err)
			return fmt.Sprintf("%v", err), err
		}
		return "ok", nil
	case "delMIBModule":
		if err := delMIBModule(&m); err != nil {
			astiLogger.Error(err)
			return fmt.Sprintf("%v", err), err
		}
		return "ok", nil
	case "resetArpTable":
		resetArpTable()
		return "ok", nil
	case "clearAllReport":
		clearAllReport()
		return "ok", nil
	case "clearAllAIMoldes":
		if err := bootstrap.SendMessage(aiWindow, "clearAllAIMoldes", ""); err != nil {
			astiLogger.Errorf("sendSendMessage clearAllAIMoldes error=%v", err)
		}
		return "ok", nil
	case "mapConf":
		return updateMapConf(&m)
	case "notifyConf":
		return updateNotifyConf(&m)
	case "notifyTest":
		return doNotify(&m)
	case "startDiscover":
		return doStartDiscover(&m)
	case "getDiscover":
		return struct {
			Conf discoverConfEnt
			Stat discoverStatEnt
		}{
			Conf: discoverConf,
			Stat: discoverStat,
		}, nil
	case "stopDiscover":
		go stopDiscover()
		return "ok", nil
	case "saveNode":
		return doSaveNode(&m)
	case "updateNode":
		return doUpdateNode(&m)
	case "deleteNode":
		return doDeleteNode(&m)
	case "dupNode":
		return doDupNode(&m)
	case "saveLine":
		return doSaveLine(&m)
	case "deleteLine":
		return doDeleteLine(&m)
	case "getLine":
		return doGetLine(&m)
	case "showNodeInfo", "showNodeLog", "showPolling":
		return doShowNodeWindow(&m)
	case "pollNow":
		return doPollNow(&m)
	case "showMIB":
		return doShowMIB(&m)
	case "logDisp":
		return doShowLogWindow(&m)
	case "reportDisp":
		return doShowReportWindow(&m)
	case "showPollingList":
		return doShowPollingListWindow(&m)
	case "checkAllPoll":
		checkAllPoll()
		return "ok", nil
	case "openUrl":
		return openURL(&m)
	case "setWindowInfo":
		return setWindowInfo(&m)
	}
	return "ok", nil
}

func updateMapConf(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		if err := json.Unmarshal(m.Payload, &mapConf); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := saveMapConfToDB(); err != nil {
			astiLogger.Errorf("saveMapConfToDB  error=%v", err)
			return "ng", err
		}
		updateBackImg()
		openGeoIP()
		loadGrokMap()
	}
	addEventLog(eventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   "",
		NodeName: "",
		Event:    "MAP設定を更新",
	})
	return "ok", nil
}

func updateNotifyConf(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		if err := json.Unmarshal(m.Payload, &notifyConf); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := saveNotifyConfToDB(); err != nil {
			astiLogger.Errorf("saveNotifyConfToDB  error=%v", err)
			return "ng", err
		}
	}
	addEventLog(eventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   "",
		NodeName: "",
		Event:    "通知設定を更新",
	})
	return "ok", nil
}

func doNotify(m *bootstrap.MessageIn) (interface{}, error) {
	var notifyTestConf notifyConfEnt
	if len(m.Payload) > 0 {
		if err := json.Unmarshal(m.Payload, &notifyTestConf); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := sendTestMail(&notifyTestConf); err != nil {
			astiLogger.Errorf("sendTestMail  error=%v", err)
			return fmt.Sprintf("%v", err), nil
		}
	}
	return "ok", nil
}

func doStartDiscover(m *bootstrap.MessageIn) (interface{}, error) {
	if discoverStat.Running {
		return "ng", nil
	}
	if len(m.Payload) > 0 {
		if err := json.Unmarshal(m.Payload, &discoverConf); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := saveDiscoverConfToDB(); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		startDiscover()
	}
	return "ok", nil
}

func doSaveNode(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var n nodeEnt
		if err := json.Unmarshal(m.Payload, &n); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if n.ID == "" {
			if err := addNode(&n); err != nil {
				astiLogger.Errorf("addNode %s error=%v", m.Name, err)
				return "ng", err
			}
		} else {
			ntmp, ok := nodes[n.ID]
			if !ok {
				astiLogger.Errorf("%s invalid nodeid %s", m.Name, n.ID)
				return "ng", errInvalidNode
			}
			ntmp.Community = n.Community
			ntmp.Descr = n.Descr
			ntmp.IP = n.IP
			ntmp.Icon = n.Icon
			ntmp.Name = n.Name
			ntmp.URL = n.URL
			if err := updateNode(ntmp); err != nil {
				astiLogger.Errorf("editNode %s error=%v", m.Name, err)
				return "ng", err
			}
			addEventLog(eventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   n.ID,
				NodeName: n.Name,
				Event:    "ノード設定を更新",
			})
		}
	}
	applyMapData()
	return "ok", nil
}

func doUpdateNode(m *bootstrap.MessageIn) (interface{}, error) {
	var n nodeEnt
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &n); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	ntmp, ok := nodes[n.ID]
	if !ok {
		astiLogger.Errorf("%s  invalid node %s", m.Name, n.ID)
		return "ng", errInvalidNode
	}
	ntmp.X = n.X
	ntmp.Y = n.Y
	if err := updateNode(ntmp); err != nil {
		astiLogger.Errorf("editNode %s error=%v", m.Name, err)
		return "ng", err
	}
	return "ok", nil
}

func doDeleteNode(m *bootstrap.MessageIn) (interface{}, error) {
	var nodeID string
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	if _, ok := nodes[nodeID]; !ok {
		return "ng", errInvalidNode
	}
	name := nodes[nodeID].Name
	if err := deleteNode(nodeID); err != nil {
		return "ng", err
	}
	addEventLog(eventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   nodeID,
		NodeName: name,
		Event:    "ノード削除",
	})
	return "ok", nil
}

func doDupNode(m *bootstrap.MessageIn) (interface{}, error) {
	var nodeID string
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
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
		astiLogger.Errorf("dupNode %s error=%v", m.Name, err)
		return "ng", err
	}
	addEventLog(eventLogEnt{
		Type:     "user",
		Level:    "info",
		NodeID:   n.ID,
		NodeName: n.Name,
		Event:    "ノード複製",
	})
	return n, nil
}

func doSaveLine(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var l lineEnt
		if err := json.Unmarshal(m.Payload, &l); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if l.ID == "" {
			if err := addLine(&l); err != nil {
				astiLogger.Errorf("addLine %s error=%v", m.Name, err)
				return "ng", err
			}
		} else {
			if err := updateLine(&l); err != nil {
				astiLogger.Errorf("updateLine %s error=%v", m.Name, err)
				return "ng", err
			}
		}
		if n, ok := nodes[l.NodeID1]; ok {
			addEventLog(eventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   l.NodeID1,
				NodeName: n.Name,
				Event:    "ライン更新",
			})
		}
		if n, ok := nodes[l.NodeID2]; ok {
			addEventLog(eventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   l.NodeID2,
				NodeName: n.Name,
				Event:    "ライン更新",
			})
		}
		updateLineState()
	}
	applyMapData()
	return "ok", nil
}

func doDeleteLine(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var l lineEnt
		if err := json.Unmarshal(m.Payload, &l); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if l.ID == "" {
			astiLogger.Errorf("delLine %s ", m.Name)
			return "ng", errInvalidID
		}
		if err := deleteLine(l.ID); err != nil {
			astiLogger.Errorf("deleteLine %s error=%v", m.Name, err)
			return "ng", err
		}
		if n, ok := nodes[l.NodeID1]; ok {
			addEventLog(eventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   l.NodeID1,
				NodeName: n.Name,
				Event:    "ライン削除",
			})
		}
		if n, ok := nodes[l.NodeID2]; ok {
			addEventLog(eventLogEnt{
				Type:     "user",
				Level:    "info",
				NodeID:   l.NodeID2,
				NodeName: n.Name,
				Event:    "ライン削除",
			})
		}
	}
	go applyMapData()
	return "ok", nil
}

func doGetLine(m *bootstrap.MessageIn) (interface{}, error) {
	var l lineEnt
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &l); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	n1, ok := nodes[l.NodeID1]
	if !ok {
		astiLogger.Errorf("Invalid Node %s", m.Name)
		return "ng", errInvalidNode
	}
	n2, ok := nodes[l.NodeID2]
	if !ok {
		astiLogger.Errorf("Invalid Node %s", m.Name)
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
	pollings.Range(func(_, p interface{}) bool {
		if p.(*pollingEnt).NodeID == l.NodeID1 {
			dlgParam.Pollings1 = append(dlgParam.Pollings1, *p.(*pollingEnt))
		}
		if p.(*pollingEnt).NodeID == l.NodeID2 {
			dlgParam.Pollings2 = append(dlgParam.Pollings2, *p.(*pollingEnt))
		}
		return true
	})
	return dlgParam, nil
}

func doShowNodeWindow(m *bootstrap.MessageIn) (interface{}, error) {
	var nodeID string
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	if err := bootstrap.SendMessage(nodeWindow, "setNodeID", nodeID); err != nil {
		astiLogger.Errorf("sendSendMessage %s error=%v", m.Name, err)
		return "ng", err
	}
	if err := bootstrap.SendMessage(nodeWindow, "setMode", m.Name); err != nil {
		astiLogger.Errorf("sendSendMessage %s error=%v", m.Name, err)
		return "ng", err
	}
	nodeWindow.Show()
	return "ok", nil
}

func doPollNow(m *bootstrap.MessageIn) (interface{}, error) {
	var nodeID string
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	pollNowNode(nodeID)
	return "ok", nil
}

func setWindowInfo(m *bootstrap.MessageIn) (interface{}, error) {
	var wi windowInfoEnt
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &wi); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	mainWindowInfo.Top = wi.Top
	mainWindowInfo.Left = wi.Left
	mainWindowInfo.Width = wi.Width
	mainWindowInfo.Height = wi.Height
	saveMainWindowInfoToDB()
	return "ok", nil
}

func doShowMIB(m *bootstrap.MessageIn) (interface{}, error) {
	var nodeID string
	if len(m.Payload) < 1 {
		return "ng", errNoPayload
	}
	if err := json.Unmarshal(m.Payload, &nodeID); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	params := struct {
		NodeID   string
		NodeName string
		MibNames []string
	}{
		NodeID:   nodeID,
		MibNames: mib.GetNameList(),
	}
	n, ok := nodes[nodeID]
	if !ok {
		astiLogger.Errorf("showMIB Invalid nodID %s", nodeID)
		return "ng", nil
	}
	params.NodeName = n.Name
	if err := bootstrap.SendMessage(mibWindow, "setParams", params); err != nil {
		astiLogger.Errorf("sendSendMessage %s error=%v", m.Name, err)
		return "ng", err
	}
	mibWindow.Show()
	return "ok", nil
}

func doShowLogWindow(m *bootstrap.MessageIn) (interface{}, error) {
	if err := bootstrap.SendMessage(logWindow, "show", ""); err != nil {
		astiLogger.Errorf("sendSendMessage %s error=%v", m.Name, err)
		return "ng", err
	}
	logWindow.Show()
	return "ok", nil
}

func doShowReportWindow(m *bootstrap.MessageIn) (interface{}, error) {
	if err := bootstrap.SendMessage(reportWindow, "show", ""); err != nil {
		astiLogger.Errorf("sendSendMessage %s error=%v", m.Name, err)
		return "ng", err
	}
	reportWindow.Show()
	return "ok", nil
}

func doShowPollingListWindow(m *bootstrap.MessageIn) (interface{}, error) {
	if err := bootstrap.SendMessage(pollingListWindow, "show", ""); err != nil {
		astiLogger.Errorf("sendSendMessage %s error=%v", m.Name, err)
		return "ng", err
	}
	pollingListWindow.Show()
	return "ok", nil
}

func applyMapConf() {
	if err := bootstrap.SendMessage(mainWindow, "mapConf", mapConf); err != nil {
		astiLogger.Errorf("sendSendMessage mapConf error=%v", err)
		return
	}
	if err := bootstrap.SendMessage(mainWindow, "notifyConf", notifyConf); err != nil {
		astiLogger.Errorf("sendSendMessage notifyConf error=%v", err)
		return
	}
}

func applyMapData() {
	if err := bootstrap.SendMessage(mainWindow, "nodes", nodes); err != nil {
		astiLogger.Errorf("sendSendMessage nodes error=%v", err)
		return
	}
	if err := bootstrap.SendMessage(mainWindow, "lines", lines); err != nil {
		astiLogger.Errorf("sendSendMessage lines error=%v", err)
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
	updateBackImg()
	applyMapConf()
	applyMapData()
	timer := time.NewTicker(time.Second * 10)
	i := 6
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case p := <-pollingStateChangeCh:
			stateCheckNodes[p.NodeID] = true
		case <-timer.C:
			doPollingCh <- true
			lastLog = sendLogs(lastLog)
			if len(stateCheckNodes) > 0 {
				astiLogger.Infof("State Change Nodes %d", len(stateCheckNodes))
				for k := range stateCheckNodes {
					updateNodeState(k)
					delete(stateCheckNodes, k)
				}
				updateLineState()
				applyMapData()
			}
			i++
			if i > 5 {
				updateDBStats()
				i = 0
				if err := bootstrap.SendMessage(mainWindow, "dbStats", dbStats); err != nil {
					astiLogger.Errorf("sendSendMessage dbStats error=%v", err)
					return
				}
			}
		}
	}
}

func sendLogs(lastLog string) string {
	list := getEventLogList(lastLog, mapConf.LogDispSize)
	if len(list) > 0 {
		astiLogger.Infof("Send Logs %d", len(list))
		if err := bootstrap.SendMessage(mainWindow, "logs", list); err != nil {
			astiLogger.Errorf("sendSendMessage logs error=%v", err)
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
	pollings.Range(func(_, p interface{}) bool {
		if p.(*pollingEnt).NodeID != nodeID {
			return true
		}
		s := p.(*pollingEnt).State
		if s == "high" {
			n.State = "high"
			return false
		}
		if s == "low" {
			n.State = "low"
			return true
		}
		if n.State == "low" {
			return true
		}
		if s == "repair" {
			n.State = "repair"
		}
		if n.State == "repair" || n.State != "unkown" {
			return true
		}
		n.State = s
		return true
	})
}

func updateLineState() {
	for _, l := range lines {
		if p, ok := pollings.Load(l.PollingID1); ok {
			l.State1 = p.(*pollingEnt).State
		}
		if p, ok := pollings.Load(l.PollingID2); ok {
			l.State2 = p.(*pollingEnt).State
		}
	}
}

func pollNowNode(nodeID string) {
	nodeName := "Unknown"
	if n, ok := nodes[nodeID]; ok {
		nodeName = n.Name
	}
	updateList := []*pollingEnt{}
	pollings.Range(func(_, v interface{}) bool {
		p := v.(*pollingEnt)
		if p.NodeID == nodeID && p.State != "normal" {
			p.State = "unkown"
			p.NextTime = 0
			pollingStateChangeCh <- p
			addEventLog(eventLogEnt{
				Type:     "user",
				Level:    p.State,
				NodeID:   p.NodeID,
				NodeName: nodeName,
				Event:    "ポーリング再確認:" + p.Name,
			})
			updateList = append(updateList, p)
		}
		return true
	})
	for _, p := range updateList {
		updatePolling(p)
	}
	doPollingCh <- true
}

func checkAllPoll() {
	updateList := []*pollingEnt{}
	pollings.Range(func(_, v interface{}) bool {
		p := v.(*pollingEnt)
		if p.State != "normal" {
			p.State = "unkown"
			p.NextTime = 0
			nodeName := "Unknown"
			if n, ok := nodes[p.NodeID]; ok {
				nodeName = n.Name
			}
			pollingStateChangeCh <- p
			addEventLog(eventLogEnt{
				Type:     "user",
				Level:    p.State,
				NodeID:   p.NodeID,
				NodeName: nodeName,
				Event:    "ポーリング再確認:" + p.Name,
			})
			updateList = append(updateList, p)
		}
		return true
	})
	for _, p := range updateList {
		updatePolling(p)
	}
	doPollingCh <- true
}

func updateBackImg() {
	path := filepath.Join(app.Paths().DataDirectory(), "resources", "app", "images", "backimg")
	if mapConf.BackImg != "" {
		os.Remove(path)
		src, err := os.Open(mapConf.BackImg)
		if err != nil {
			astiLogger.Errorf("updateBackImg err=%v", err)
			return
		}
		defer src.Close()
		dst, err := os.Create(path)
		if err != nil {
			astiLogger.Errorf("updateBackImg err=%v", err)
			return
		}
		defer dst.Close()
		io.Copy(dst, src)
	} else {
		os.Remove(path)
	}
}
