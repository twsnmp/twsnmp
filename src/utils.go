package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

func openURL(m *bootstrap.MessageIn) (interface{}, error) {
	if len(m.Payload) < 1 {
		astiLogger.Errorf("openUrl no payload")
		return "ng", nil
	}
	var url string
	if err := json.Unmarshal(m.Payload, &url); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
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
		astiLogger.Errorf("openUrl err=%v", err)
		return "ng", err
	}
	return "ok", err
}
