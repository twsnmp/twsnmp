package main

import (
	"encoding/json"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)


// pollingMessageHandler handles messages
func pollingMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "close":
			pollingWindow.Hide()
			return "ok",nil
		case "get":
			if len(m.Payload) > 0 {
				var param = struct {
					StartTime string
					EndTime   string
					PollingID  string
				}{}
				if err := json.Unmarshal(m.Payload, &param); err != nil {
					astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
					return []pollingLogEnt{}, errInvalidParams
				}
				return getPollingLog(param.StartTime,param.EndTime,param.PollingID),nil
			}
			return []pollingLogEnt{}, errInvalidParams
	}
	astilog.Errorf("Unknow Message Name=%s",m.Name)
	return "ok",nil
}
