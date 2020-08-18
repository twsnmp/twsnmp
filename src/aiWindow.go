package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
)

var aiBusy = false

const yasumi = `date,name
2019-01-01,元日
2019-01-14,成人の日
2019-02-11,建国記念の日
2019-03-21,春分の日
2019-04-29,昭和の日
2019-04-30,休日
2019-05-01,天皇の即位の日
2019-05-02,休日
2019-05-03,憲法記念日
2019-05-04,みどりの日
2019-05-05,こどもの日
2019-05-06,休日
2019-07-15,海の日
2019-08-11,山の日
2019-08-12,休日
2019-09-16,敬老の日
2019-09-23,秋分の日
2019-10-14,体育の日
2019-11-03,文化の日
2019-11-23,勤労感謝の日
2019-12-30,冬季休業
2019-12-31,冬季休業
2020-01-01,元日
2020-01-02,冬季休業
2020-01-03,冬季休業
2020-01-13,成人の日
2020-02-11,建国記念の日
2020-02-23,天皇の即位の日
2020-02-24,休日
2020-03-20,春分の日
2020-04-29,昭和の日
2020-05-03,憲法記念日
2020-05-04,みどりの日
2020-05-05,こどもの日
2020-05-06,休日
2020-07-23,海の日
2020-07-24,スポーツの日
2020-08-10,山の日
2020-09-21,敬老の日
2020-09-22,秋分の日
2020-11-03,文化の日
2020-11-23,勤労感謝の日
`

var yasumiMap = make(map[string]bool)

// aiMessageHandler handles messages
func aiMessageHandler(w *astilectron.Window, m bootstrap.MessageIn) (interface{}, error) {
	switch m.Name {
	case "done":
		aiBusy = false
		doneAI(&m)
		return "ok", nil
	}
	return "ok", nil
}

func aiWindowBackend(ctx context.Context) {
	makeYasumiMap()
	timer := time.NewTicker(time.Second * 10)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			if aiBusy {
				continue
			}
			aiBusy = checkAI()
		}
	}
}

type aiReq struct {
	PollingID string
	TimeStamp []int64
	Data      [][]float64
}

func makeYasumiMap() {
	for _, l := range strings.Split(yasumi, "\n") {
		y := strings.Split(l, ",")
		if len(y) == 2 {
			if _, err := time.Parse("2006-01-02", y[0]); err == nil {
				yasumiMap[y[0]] = true
			}
		}
	}
}

var checkAIMap = make(map[string]int64)

func checkAI() bool {
	pollings.Range(func(_, v interface{}) bool {
		p := v.(*pollingEnt)
		if p.LogMode == logModeAI {
			if _, ok := checkAIMap[p.ID]; !ok {
				checkAIMap[p.ID] = 0
			}
		}
		return true
	})
	now := time.Now().Unix()
	var selPolling *pollingEnt
	delList := []string{}
	for id, n := range checkAIMap {
		v, ok := pollings.Load(id)
		if !ok || v == nil {
			delList = append(delList, id)
			continue
		}
		p := v.(*pollingEnt)
		if n > now {
			continue
		}
		if selPolling == nil || checkAIMap[selPolling.ID] > n {
			selPolling = p
		}
	}
	for _, id := range delList {
		astiLogger.Debugf("checkAIMap delete %s", id)
		delete(checkAIMap, id)
	}
	if selPolling == nil {
		return false
	}
	checkAIMap[selPolling.ID] = now + 60*5
	return doAI(selPolling)
}

func resetAIResult(id string) {
	deleteAIReesult(id)
	checkAIMap[id] = 0
}

func checkLastAIResultTime(id string) bool {
	last, err := loadAIReesult(id)
	if err != nil {
		astiLogger.Errorf("loadAIReesult  id=%s err=%v", id, err)
		deleteAIReesult(id)
		return true
	}
	astiLogger.Debugf("checkLastAIResultTime %s = %d diff=%d", id, last.LastTime, time.Now().Unix()-last.LastTime)
	if last.LastTime < time.Now().Unix()-60*60 {
		return true
	}
	return false
}

func doAI(p *pollingEnt) bool {
	if !checkLastAIResultTime(p.ID) {
		return false
	}
	req := &aiReq{
		PollingID: p.ID,
	}
	if p.Type == "syslogpri" {
		makeAIDataFromSyslogPriPolling(req)
	} else {
		makeAIDataFromPolling(req)
	}
	if len(req.Data) < 10 {
		astiLogger.Debugf("doAI Skip No data %s %s %v", p.ID, p.Name, req)
		return false
	}
	astiLogger.Infof("doAI Start %s %s %d", p.ID, p.Name, len(req.Data))
	if err := bootstrap.SendMessage(aiWindow, "doAI", req); err != nil {
		astiLogger.Errorf("sendSendMessage doAI error=%v", err)
		return false
	}
	return true
}

func makeAIDataFromSyslogPriPolling(req *aiReq) {
	logs := getAllPollingLog(req.PollingID)
	if len(logs) < 1 {
		return
	}
	st := 3600 * (time.Unix(0, logs[0].Time).Unix() / 3600)
	ent := make([]float64, 257)
	var maxVal float64
	for _, l := range logs {
		ct := 3600 * (time.Unix(0, l.Time).Unix() / 3600)
		if st != ct {
			ts := time.Unix(ct, 0)
			ent[0] = float64(ts.Hour()) / 24.0
			if _, ok := yasumiMap[ts.Format("2006-01-02")]; ok {
				ent[1] = 0.0
			} else {
				ent[1] = float64(ts.Weekday()) / 6.0
			}
			req.TimeStamp = append(req.TimeStamp, ts.Unix())
			req.Data = append(req.Data, ent)
			ent = make([]float64, 257)
			st = ct
		}
		for _, e := range strings.Split(l.StrVal, ";") {
			var pri int
			var count int
			if n, err := fmt.Sscanf(e, "%d=%d", &pri, &count); err == nil && n == 2 {
				if pri >= 0 && pri < 256 {
					ent[pri+2] += float64(count)
					if maxVal < ent[pri+2] {
						maxVal = ent[pri+2]
					}
				}
			}
		}
	}
	if maxVal == 0.0 {
		return
	}
	for i := range req.Data {
		for j := range req.Data[i] {
			if j < 2 {
				continue
			}
			req.Data[i][j] /= maxVal
		}
	}
	return
}

const entLen = 20

func makeAIDataFromPolling(req *aiReq) {
	logs := getAllPollingLog(req.PollingID)
	if len(logs) < 1 {
		return
	}
	st := 3600 * (time.Unix(0, logs[0].Time).Unix() / 3600)
	ent := make([]float64, entLen)
	maxVals := make([]float64, entLen)
	var count float64
	for _, l := range logs {
		ct := 3600 * (time.Unix(0, l.Time).Unix() / 3600)
		if st != ct {
			ts := time.Unix(ct, 0)
			ent[0] = float64(ts.Hour())
			if _, ok := yasumiMap[ts.Format("2006-01-02")]; ok {
				ent[1] = 0.0
			} else {
				ent[1] = float64(ts.Weekday())
			}
			if count == 0.0 {
				continue
			}
			for i := 0; i < len(ent); i++ {
				if i >= 4 {
					ent[i] /= count
				}
				if maxVals[i] < ent[i] {
					maxVals[i] = ent[i]
				}
			}
			req.TimeStamp = append(req.TimeStamp, ts.Unix())
			req.Data = append(req.Data, ent)
			ent = make([]float64, entLen)
			st = ct
			count = 0.0
		}
		count += 1.0
		if l.State == "normal" || l.State == "repair" {
			if ent[2] < l.NumVal {
				ent[2] = l.NumVal
			}
			if ent[3] == 0.0 || l.NumVal < ent[3] {
				ent[3] = l.NumVal
			}
			ent[4] += float64(l.NumVal)
		}
		ent[5] += getStateNum(l.State)
		lr := make(map[string]string)
		if err := json.Unmarshal([]byte(l.StrVal), &lr); err == nil {
			i := 6
			keys := make([]string, 0, len(lr))
			for k := range lr {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				if fv, err := strconv.ParseFloat(lr[k], 64); err == nil {
					ent[i] += fv
					i++
				}
				if i >= len(ent) {
					break
				}
			}
		}
	}
	for i := range req.Data {
		for j := range req.Data[i] {
			if maxVals[j] > 0.0 {
				req.Data[i][j] /= maxVals[j]
			} else {
				req.Data[i][j] = 0.0
			}
		}
	}
	return
}

func getStateNum(s string) float64 {
	if s == "repair" || s == "normal" {
		return 1.0
	}
	if s == "unknown" {
		return 0.5
	}
	return 0.0
}

func doneAI(m *bootstrap.MessageIn) {
	if len(m.Payload) < 1 {
		astiLogger.Errorf("sendSendMessage doneAI Payload len=0")
		return
	}
	var res aiResult
	if err := json.Unmarshal(m.Payload, &res); err != nil {
		astiLogger.Errorf("Unmarshal %s error=%v", m.Name, err)
		return
	}
	if len(res.ScoreData) < 1 {
		return
	}
	if err := saveAIResultToDB(&res); err != nil {
		astiLogger.Errorf("saveAIResultToDB err=%v", err)
	}
	var p *pollingEnt
	if v, ok := pollings.Load(res.PollingID); ok {
		p = v.(*pollingEnt)
	}
	if p == nil {
		return
	}
	astiLogger.Infof("doneAI pollingID=%s pollingName=%s iffluxdb=%s lenScore=%d",
		res.PollingID, p.Name, influxdbConf.AIScore, len(res.ScoreData))
	if len(res.ScoreData) > 0 {
		ls := res.ScoreData[len(res.ScoreData)-1][1]
		if ls > float64(mapConf.AIThreshold) {
			nodeName := "Unknown"
			if n, ok := nodes[p.NodeID]; ok {
				nodeName = n.Name
			}
			addEventLog(eventLogEnt{
				Type:     "ai",
				Level:    mapConf.AILevel,
				NodeID:   p.NodeID,
				NodeName: nodeName,
				Event:    fmt.Sprintf("AI分析レポート:%s(%s):%f", p.Name, p.Type, ls),
			})
		}
		if influxdbConf.AIScore == "send" {
			if err := sendAIScoreToInfluxdb(p, &res); err != nil {
				astiLogger.Errorf("sendAIScoreToInfluxdb err=%v", err)
			}
		}
	}
	return
}
