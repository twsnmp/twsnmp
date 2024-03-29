package main

import (
	"encoding/json"
	"fmt"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

type filterEnt struct {
	StartTime string
	EndTime   string
	Filter    string
	LogType   string
}

type arpEnt struct {
	IP     string
	MAC    string
	Vendor string
}

type arpResEnt struct {
	Arps []arpEnt
	Logs []logEnt
}

// logMessageHandler handles messages
func logMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		_ = logWindow.Hide()
		_ = logWindow.CloseDevTools()
		return "ok", nil
	case "searchLog":
		return searchLog(&m)
	case "showPolling":
		return showPolling(&m)
	case "savePolling":
		return savePolling(&m)
	case "deletePolling":
		return deletePollingMsg(&m)
	case "pollNow":
		return pollNow(&m)
	}
	astiLogger.Errorf("Unknow Message Name=%s", m.Name)
	return "ok", nil
}

func searchLog(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var filter filterEnt
		if err := json.Unmarshal(m.Payload, &filter); err != nil {
			astiLogger.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return []eventLogEnt{}, errInvalidParams
		}
		if filter.LogType == "log" {
			return getEventLogs(&filter), nil
		}
		if filter.LogType == "arp" {
			return getArpRes(&filter)
		}
		return getLogs(&filter), nil
	}
	return []eventLogEnt{}, errInvalidParams
}

func getArpRes(filter *filterEnt) (arpResEnt, error) {
	arps := []arpEnt{}
	for ip, mac := range arpTable {
		arps = append(arps, arpEnt{
			IP:     ip,
			MAC:    mac,
			Vendor: oui.Find(mac),
		})
	}
	filter.LogType = "arplog"
	logs := getLogs(filter)
	return arpResEnt{
		Arps: arps,
		Logs: logs,
	}, nil
}
