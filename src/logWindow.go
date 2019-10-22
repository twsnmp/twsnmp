package main

import (
	"encoding/json"
	"fmt"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)

type filterEnt struct {
	StartTime string
	EndTime   string
	Filter    string
	LogType   string
}

type logCountEnt struct {
	High int32
	Low  int32
	Warn int32
	Normal int32
	Other int32
	Time int64
}

// logMessageHandler handles messages
func logMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "close":
			logWindow.Hide()
			logWindow.CloseDevTools()
			return "ok",nil
		case "getLogPollings":
			return getLogPollings(),nil
		case "getNodes":
			return getNodes(),nil
		case "searchLog":
			return searchLog(&m)
		case "savePolling":
			return savePolling(&m)
		case "deletePolling":
			return deletePollingMsg(&m)
		case "pollNow":
			return pollNow(&m)
	}
	astilog.Errorf("Unknow Message Name=%s",m.Name)
	return "ok",nil
}

func searchLog(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var filter filterEnt
		if err := json.Unmarshal(m.Payload, &filter); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
			return []eventLogEnt{}, errInvalidParams
		}
		if filter.LogType == "log"{
			return getEventLogs(&filter),nil
		}
		return getLogs(&filter),nil
	}
	return []eventLogEnt{}, errInvalidParams
}

func getNodes() map[string]string{
	ret := map[string]string{}
	ret[""] = ""
	for k,n := range nodes {
		ret[n.Name] = k
	}
	return ret
}