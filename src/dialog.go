package main

import (
	"fmt"
	"encoding/json"
	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
)


// dialogMessageHandler handles messages
func dialogMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{},error) {
	switch m.Name {
		case "cancel":{
			dialogWindow.Hide()
		}
		case "save.configMap":{
			if len(m.Payload) > 0 {
				if err := json.Unmarshal(m.Payload, &mapConf); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return err.Error(),err
				}
				if err := saveMapConfToDB(); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return err.Error(),err
				}
				applyMapConf()
				dialogWindow.Hide()
			}
			return "",nil
		}
		case "startDiscover":{
			if len(m.Payload) > 0 {
				if err := json.Unmarshal(m.Payload, &discoverConf); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return err.Error(),err
				}
				if err := saveDiscoverConfToDB(); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return err.Error(),err
				}
				startDiscover()
				dialogWindow.Hide()
			}
			return "",nil
		}
		case "stopDiscover":{
			stopDiscover()
			dialogWindow.Hide()
			return "",nil
		}
		case "save.editNode": {
			if len(m.Payload) > 0 {
				var n nodeEnt
				if err := json.Unmarshal(m.Payload, &n); err != nil {
					astilog.Errorf("Unmarshal %s error=%v",m.Name, err)
					return "ng",err
				}
				if n.ID == "" {
					if err := addNode(&n); err != nil {
						astilog.Errorf("addNode %s error=%v",m.Name, err)
						return "ng",err
					}
				} else {
					ntmp,ok := nodes[n.ID]
					if !ok {
						astilog.Errorf("%s invalid nodeid %s",m.Name, n.ID)
						return "ng",errInvalidNode
					}
					ntmp.Community = n.Community 
					ntmp.Descr = n.Descr 
					ntmp.IP = n.IP 
					ntmp.Icon = n.Icon 
					ntmp.Name = n.Name 
					if err := updateNode(ntmp);err != nil {
						astilog.Errorf("editNode %s error=%v",m.Name, err)
						return "ng",err
					}
				}
			}		
			applyMapData()
			dialogWindow.Hide()
			return "",nil
		}
		case "save.editLine": {
			if len(m.Payload) > 0 {
				var l lineEnt
				if err := json.Unmarshal(m.Payload, &l); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return err.Error(),err
				}
				if l.ID == "" {
					if err := addLine(&l); err != nil {
						astilog.Error(fmt.Sprintf("addLine %s error=%v",m.Name, err))
						return err.Error(),err
					}
				} else {
					if err := updateLine(&l);err != nil {
						astilog.Error(fmt.Sprintf("updateLine %s error=%v",m.Name, err))
						return err.Error(),err
					}
				}
				updateLineState()
			}		
			applyMapData()
			dialogWindow.Hide()
			return "",nil
		}
		case "del.editLine": {
			if len(m.Payload) > 0 {
				var l lineEnt
				if err := json.Unmarshal(m.Payload, &l); err != nil {
					astilog.Error(fmt.Sprintf("Unmarshal %s error=%v",m.Name, err))
					return err.Error(),err
				}
				if l.ID == "" {
					astilog.Error(fmt.Sprintf("delLine %s ",m.Name))
					return "ng",errInvalidID
				}
				if err := deleteLine(l.ID);err != nil {
					astilog.Error(fmt.Sprintf("deleteLine %s error=%v",m.Name, err))
					return err.Error(),err
				}
			}		
			applyMapData()
			dialogWindow.Hide()
			return "",nil
		}
	}
	return "",nil
}
