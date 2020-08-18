package main

import (
	"encoding/json"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

// pollingMessageHandler handles messages
func pollingMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		pollingWindow.Hide()
		return "ok", nil
	case "clear":
		if len(m.Payload) > 0 {
			var pollingID string
			if err := json.Unmarshal(m.Payload, &pollingID); err != nil {
				astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
				return "ng", errInvalidParams
			}
			if err := clearPollingLog(pollingID); err != nil {
				return "ng", err
			}
			if err := bootstrap.SendMessage(aiWindow, "deleteModel", pollingID); err != nil {
				astiLogger.Errorf("sendSendMessage deleteModel error=%v", err)
			}
			return "ok", nil
		}
		return "ng", errInvalidParams
	case "getai":
		if len(m.Payload) > 0 {
			var id string
			if err := json.Unmarshal(m.Payload, &id); err != nil {
				astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
				return &aiResult{}, errInvalidParams
			}
			return loadAIReesult(id)
		}
		return &aiResult{}, errInvalidParams
	case "doai":
		if len(m.Payload) > 0 {
			var id string
			if err := json.Unmarshal(m.Payload, &id); err != nil {
				astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
				return "ng", errInvalidParams
			}
			resetAIResult(id)
			return "ok", nil
		}
		return "ng", errInvalidParams
	case "get":
		if len(m.Payload) > 0 {
			var param = struct {
				StartTime string
				EndTime   string
				PollingID string
			}{}
			if err := json.Unmarshal(m.Payload, &param); err != nil {
				astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
				return []pollingLogEnt{}, errInvalidParams
			}
			return getPollingLog(param.StartTime, param.EndTime, param.PollingID), nil
		}
		return []pollingLogEnt{}, errInvalidParams
	}
	astiLogger.Errorf("Unknow Message Name=%s", m.Name)
	return "ok", nil
}
