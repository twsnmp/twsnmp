package main

import (

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
	}
	astilog.Errorf("Unknow Message Name=%s",m.Name)
	return "ok",nil
}
