package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
)

var influxc client.Client
var muInfluxc sync.Mutex

func setupInfluxdb() error {
	closeInfluxdb()
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxdbConf.URL == "" {
		return nil
	}
	var err error
	conf := client.HTTPConfig{
		Addr:               influxdbConf.URL,
		Timeout:            time.Second * 5,
		InsecureSkipVerify: true,
	}
	if influxdbConf.User != "" && influxdbConf.Password != "" {
		conf.Username = influxdbConf.User
		conf.Password = influxdbConf.Password
	}
	influxc, err = client.NewHTTPClient(conf)
	if err != nil {
		influxc = nil
		return err
	}
	return checkInfluxdb()
}

func checkInfluxdb() error {
	q := client.NewQuery("SHOW DATABASES", "", "")
	if response, err := influxc.Query(q); err == nil && response.Error() == nil {
		for _, r := range response.Results {
			for _, s := range r.Series {
				for _, ns := range s.Values {
					for _, n := range ns {
						if name, ok := n.(string); ok {
							if name == influxdbConf.DB {
								return nil
							}
						}
					}
				}
			}
		}
	} else {
		return err
	}
	qs := fmt.Sprintf(`CREATE DATABASE "%s"`, influxdbConf.DB)
	if influxdbConf.Duration != "" {
		qs += " WITH DURATION " + influxdbConf.Duration
	}
	q = client.NewQuery(qs, "", "")
	if response, err := influxc.Query(q); err != nil || response.Error() != nil {
		return err
	}
	return nil
}

func dropInfluxdb() error {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return nil
	}
	qs := fmt.Sprintf(`DROP DATABASE "%s"`, influxdbConf.DB)
	q := client.NewQuery(qs, "", "")
	if response, err := influxc.Query(q); err != nil || response.Error() != nil {
		return err
	}
	return nil
}

func sendPollingLogToInfluxdb(p *pollingEnt) error {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return nil
	}
	n, ok := nodes[p.NodeID]
	if !ok {
		return errInvalidID
	}
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  influxdbConf.DB,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{
		"map":       mapConf.MapName,
		"node":      n.Name,
		"nodeID":    n.ID,
		"pollingID": p.ID,
	}
	fields := map[string]interface{}{
		"numVal": p.LastVal,
	}
	lr := make(map[string]string)
	if err := json.Unmarshal([]byte(p.LastResult), &lr); err == nil {
		for k, v := range lr {
			if fv, err := strconv.ParseFloat(v, 64); err == nil {
				fields[k] = fv
			} else {
				fields[k] = v
			}
		}
	}
	pt, err := client.NewPoint(p.Name, tags, fields, time.Unix(0, p.LastTime))
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	if err := influxc.Write(bp); err != nil {
		return err
	}
	return nil
}

func sendAIScoreToInfluxdb(p *pollingEnt, res *aiResult) error {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return nil
	}
	n, ok := nodes[p.NodeID]
	if !ok {
		return errInvalidID
	}
	qs := fmt.Sprintf(`DROP SERIES FROM "AIScore" WHERE "pollingID" = "%s" `, p.ID)
	q := client.NewQuery(qs, influxdbConf.DB, "")
	if response, err := influxc.Query(q); err != nil {
		astiLogger.Errorf("sendAIScoreToInfluxdb err=%v", err)
		return err
	} else if response == nil {
		astiLogger.Errorf("sendAIScoreToInfluxdb err=%v resp=nil", err)
		return err
	} else if response.Error() != nil {
		astiLogger.Errorf("sendAIScoreToInfluxdb err=%v respError=%v", err, response.Error())
		return err
	}
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  influxdbConf.DB,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	// Create a point and add to batch
	tags := map[string]string{
		"map":       mapConf.MapName,
		"node":      n.Name,
		"nodeID":    n.ID,
		"pollingID": p.ID,
	}
	for _, score := range res.ScoreData {
		if len(score) < 2 {
			continue
		}
		fields := map[string]interface{}{
			"AIScore": score[1],
		}
		pt, err := client.NewPoint("AIScore", tags, fields, time.Unix(int64(score[0]), 0))
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}
	// Write the batch
	if err := influxc.Write(bp); err != nil {
		return err
	}
	return nil

}

func closeInfluxdb() {
	muInfluxc.Lock()
	defer muInfluxc.Unlock()
	if influxc == nil {
		return
	}
	influxc.Close()
	influxc = nil
}
