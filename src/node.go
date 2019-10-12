package main

import (
	"fmt"
	"encoding/json"
	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)


// nodeMessageHandler handles messages
func nodeMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "cancel":
			nodeWindow.Hide()
			return "ok",nil
		case "savePolling":
			if len(m.Payload) > 0 {
				var p pollingEnt
				if err := json.Unmarshal(m.Payload, &p); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v", m.Name, err))
					return "ng", err
				}
				if p.ID == "" {
					if err := addPolling(&p); err != nil {
						astilog.Error(fmt.Sprintf("addPolling %s error=%v", m.Name, err))
						return "ng", err
					}
				} else {
					if err := updatePolling(&p); err != nil {
						astilog.Error(fmt.Sprintf("updatePolling %s error=%v", m.Name, err))
						return "ng", err
					}
				}
			}
			return "ok", nil
		case "updatePolling":
		case "deltePolling":
	}
	return "ok",nil
}
