package main

import (
	"context"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

var aiBusy bool

// aiMessageHandler handles messages
func aiMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "close":
		return "ok", nil
	case "clear":
		return "ok", nil
	}
	return "ok", nil
}

func aiWindowBackend(ctx context.Context) {
	timer := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if aiBusy {
				continue
			}
		}
	}
}
