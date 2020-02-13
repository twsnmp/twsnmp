package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)

func openBrowser(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) < 1 {
		astilog.Errorf("openBrowser no payload")
		return "ng", nil
	}
	var url string
	if err := json.Unmarshal(m.Payload, &url); err != nil {
		astilog.Errorf("Unmarshal %s error=%v", m.Name, err)
		return "ng", err
	}
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		astilog.Errorf("openBrowser err=%v", err)
		return "ng", err
	}
	return "ok", err
}
