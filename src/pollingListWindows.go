package main

import (
	"encoding/json"
	"io/ioutil"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

// pollingListMessageHandler handles messages
func pollingListMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		_ = pollingListWindow.Hide()
		return "ok", nil
	case "getPollings":
		ret := struct {
			Pollings []pollingEnt
			Nodes    map[string]*nodeEnt
		}{
			Pollings: getPollings(),
			Nodes:    nodes,
		}
		return ret, nil
	case "savePolling":
		return savePolling(&m)
	case "showPolling":
		return showPolling(&m)
	case "deletePolling":
		return deletePollingMsg(&m)
	case "pollNow":
		return pollNow(&m)
	case "getTemplates":
		return pollingTemplates, nil
	case "saveTemplate":
		return savePollingTemplate(&m)
	case "importTemplate":
		return importPollingTemplate(&m)
	case "exportTemplate":
		return exportPollingTemplate(&m)
	case "deleteTemplate":
		return deletePollingTemplateMsg(&m)
	}
	astiLogger.Errorf("Unknow Message Name=%s", m.Name)
	return "ok", nil
}

func savePollingTemplate(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var pt pollingTemplateEnt
		if err := json.Unmarshal(m.Payload, &pt); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if pt.ID == "" {
			if err := addPollingTemplate(&pt); err != nil {
				astiLogger.Errorf("addPollingTemplate %s error=%v", m.Name, err)
				return "ng", err
			}
		} else {
			if err := updatePollingTemplate(&pt); err != nil {
				astiLogger.Errorf("updatePollingTemplate %s error=%v", m.Name, err)
				return "ng", err
			}
		}
	}
	return "ok", nil
}

func deletePollingTemplateMsg(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var id string
		if err := json.Unmarshal(m.Payload, &id); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		if err := deletePollingTemplate(id); err != nil {
			astiLogger.Errorf("deletePollingTemplate  error=%v", err)
			return "ng", err
		}
		return "ok", nil
	}
	return "ng", errInvalidNode
}

func importPollingTemplate(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var file string
		if err := json.Unmarshal(m.Payload, &file); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		js, err := ioutil.ReadFile(file)
		if err != nil {
			astiLogger.Errorf("Marshal pollingTemplates error=%v", err)
			return "ng", err
		}
		var list []pollingTemplateEnt
		if err := json.Unmarshal(js, &list); err != nil {
			astiLogger.Errorf("Unmarshal pollingTemplateEnt error=%v", err)
			return "ng", err
		}
		for i := range list {
			if err := addPollingTemplate(&list[i]); err != nil {
				astiLogger.Errorf("addPollingTemplate  error=%v", err)
			}
		}
		return "ok", nil
	}
	return "ng", errInvalidNode
}

func exportPollingTemplate(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) > 0 {
		var file string
		if err := json.Unmarshal(m.Payload, &file); err != nil {
			astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
			return "ng", err
		}
		var list []pollingTemplateEnt
		for _, val := range pollingTemplates {
			list = append(list, *val)
		}
		js, err := json.Marshal(&list)
		if err != nil {
			astiLogger.Errorf("Marshal pollingTemplates error=%v", err)
			return "ng", err
		}
		err = ioutil.WriteFile(file, js, 0644)
		if err != nil {
			astiLogger.Errorf("WriteFile pollingTemplates error=%v", err)
			return "ng", err
		}
		return "ok", nil
	}
	return "ng", errInvalidNode
}
