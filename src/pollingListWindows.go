package main

import (

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)

// pollingListMessageHandler handles messages
func pollingListMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		pollingListWindow.Hide()
		return "ok", nil
	case "getPollings":
		ret := struct {
			Pollings []pollingEnt  
			Nodes map[string]*nodeEnt
		}{
			Pollings: getPollings(),
			Nodes: nodes,
		}
		return ret,nil
	case "savePolling":
		return savePolling(&m)
	case "showPolling":
		return showPolling(&m)
	case "deletePolling":
		return deletePollingMsg(&m)
	case "pollNow":
		return pollNow(&m)
	}
	astilog.Errorf("Unknow Message Name=%s",m.Name)
	return "ok", nil
}

