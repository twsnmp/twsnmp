package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"sync"
	"time"

	astilog "github.com/asticode/go-astilog"
	"go.etcd.io/bbolt"
)

var (
	db *bbolt.DB
	// Data on Memory
	mapConf           mapConfEnt
	notifyConf        notifyConfEnt
	discoverConf      discoverConfEnt
	nodes             = make(map[string]*nodeEnt)
	lines             = make(map[string]*lineEnt)
	pollings          = sync.Map{}
	eventLogCh        = make(chan eventLogEnt, 100)
	stopEventLoggerCh = make(chan bool)
)

type nodeEnt struct {
	ID        string
	Name      string
	Descr     string
	Icon      string
	State     string
	X         int
	Y         int
	IP        string
	Community string
}

type lineEnt struct {
	ID         string
	NodeID1    string
	PollingID1 string
	State1     string
	NodeID2    string
	PollingID2 string
	State2     string
}

type pollingEnt struct {
	ID         string
	Name       string
	NodeID     string
	Type       string
	Polling    string
	Level      string
	PollInt    int
	Timeout    int
	Retry      int
	LogMode    int
	LastTime   int64
	LastResult string
	LastVal    float64
	State      string
}

type eventLogEnt struct {
	Time     int64 // UnixNano()
	Type     string
	Level    string
	NodeName string
	NodeID   string
	Event    string
}

type pollingLogEnt struct {
	Time      int64 // UnixNano()
	PollingID string
	State     string
	NumVal    float64
	StrVal    string
}

type logEnt struct {
	Time int64 // UnixNano()
	Type string
	Log  string
}

// Config Data Struct
type mapConfEnt struct {
	MapName        string
	PollInt        int
	Timeout        int
	Retry          int
	LogDays        int
	LogDispSize    int
	Community      string
	EnableSyslogd  bool
	EnableTrapd    bool
	EnableNetflowd bool
	BackImg        string
}

type notifyConfEnt struct {
	MailServer         string
	User               string
	Password           string
	InsecureSkipVerify bool
	MailTo             string
	MailFrom           string
	Subject            string
	Interval           int
	Level              string
}

type discoverConfEnt struct {
	StartIP   string
	EndIP     string
	Community string
	Timeout   int
	Retry     int
	X         int
	Y         int
}

func checkDB(path string) error {
	var err error
	d, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	defer d.Close()
	return nil
}

func openDB(path string) error {
	var err error
	db, err = bbolt.Open(path, 0600, nil)
	if err != nil {
		return err
	}
	err = initDB()
	if err != nil {
		db.Close()
		return err
	}
	err = loadConfFromDB()
	if err != nil {
		db.Close()
		return err
	}
	err = loadMapDataFromDB()
	if err != nil {
		db.Close()
		return err
	}
	return nil
}

func initDB() error {
	buckets := []string{"config", "nodes", "lines", "pollings", "logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix", "mibdb"}
	mapConf.Community = "public"
	mapConf.PollInt = 60
	mapConf.Retry = 1
	mapConf.Timeout = 1
	mapConf.LogDispSize = 5000
	mapConf.LogDays = 14
	discoverConf.Community = "public"
	discoverConf.Retry = 1
	discoverConf.Timeout = 1
	notifyConf.InsecureSkipVerify = true
	notifyConf.Interval = 60
	notifyConf.Subject = "TWSNMPからの通知"
	notifyConf.Level = "none"
	return db.Update(func(tx *bbolt.Tx) error {
		for _, b := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(b)); err != nil {
				return err
			}
		}
		return nil
	})
}

func loadConfFromDB() error {
	if db == nil {
		return errDBNotOpen
	}
	bSaveConf := false
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		v := b.Get([]byte("mapConf"))
		if v == nil {
			bSaveConf = true
			return nil
		}
		if err := json.Unmarshal(v, &mapConf); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal mapConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("discoverConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &discoverConf); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal discoverConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("notifyConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &notifyConf); err != nil {
			astilog.Error(fmt.Sprintf("Unmarshal notifyConf from DB error=%v", err))
			return err
		}
		return nil
	})
	if err == nil && bSaveConf {
		saveMapConfToDB()
		saveNotifyConfToDB()
		saveDiscoverConfToDB()
	}
	return err
}

func saveMapConfToDB() error {
	if db == nil {
		return errDBNotOpen
	}
	s, err := json.Marshal(mapConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("Bucket config is nil")
		}
		b.Put([]byte("mapConf"), s)
		return nil
	})
}

func saveNotifyConfToDB() error {
	if db == nil {
		return errDBNotOpen
	}
	s, err := json.Marshal(notifyConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("Bucket config is nil")
		}
		b.Put([]byte("notifyConf"), s)
		return nil
	})
}

func saveDiscoverConfToDB() error {
	if db == nil {
		return errDBNotOpen
	}
	s, err := json.Marshal(discoverConf)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("Bucket config is nil")
		}
		b.Put([]byte("discoverConf"), s)
		return nil
	})
}

func loadMapDataFromDB() error {
	if db == nil {
		return errDBNotOpen
	}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			var n nodeEnt
			if err := json.Unmarshal(v, &n); err == nil {
				nodes[n.ID] = &n
			}
			return nil
		})
		b = tx.Bucket([]byte("lines"))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				var l lineEnt
				if err := json.Unmarshal(v, &l); err == nil {
					lines[l.ID] = &l
				}
				return nil
			})
		}
		b = tx.Bucket([]byte("pollings"))
		if b != nil {
			b.ForEach(func(k, v []byte) error {
				var p pollingEnt
				if err := json.Unmarshal(v, &p); err == nil {
					pollings.Store(p.ID, &p)
				}
				return nil
			})
		}
		return nil
	})
	return err
}

func addNode(n *nodeEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	for {
		n.ID = makeKey()
		if _, ok := nodes[n.ID]; !ok {
			break
		}
	}
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		b.Put([]byte(n.ID), s)
		return nil
	})
	nodes[n.ID] = n
	return nil
}

func updateNode(n *nodeEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := nodes[n.ID]; !ok {
		return errInvalidID
	}
	s, err := json.Marshal(n)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		b.Put([]byte(n.ID), s)
		return nil
	})
	return nil
}

func deleteNode(nodeID string) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := nodes[nodeID]; !ok {
		return errInvalidID
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		b.Delete([]byte(nodeID))
		return nil
	})
	delete(nodes, nodeID)
	delList := []string{}
	pollings.Range(func(k, v interface{}) bool {
		if v.(*pollingEnt).NodeID == nodeID {
			delList = append(delList, k.(string))
		}
		return true
	})
	for _, k := range delList {
		deletePolling(k)
	}
	return nil
}

func findNodeFromIP(ip string) *nodeEnt {
	for _, n := range nodes {
		if n.IP == ip {
			return n
		}
	}
	return nil
}

func addLine(l *lineEnt) error {
	for {
		l.ID = makeKey()
		if _, ok := lines[l.ID]; !ok {
			break
		}
	}
	if db == nil {
		return errDBNotOpen
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		b.Put([]byte(l.ID), s)
		return nil
	})
	lines[l.ID] = l
	return nil
}

func updateLine(l *lineEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := lines[l.ID]; !ok {
		return errInvalidID
	}
	s, err := json.Marshal(l)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		b.Put([]byte(l.ID), s)
		return nil
	})
	return nil
}

func deleteLine(lineID string) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := lines[lineID]; !ok {
		return errInvalidID
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		b.Delete([]byte(lineID))
		return nil
	})
	delete(lines, lineID)
	return nil
}

// pollings
func addPolling(p *pollingEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	for {
		p.ID = makeKey()
		if _, ok := pollings.Load(p.ID); !ok {
			break
		}
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		b.Put([]byte(p.ID), s)
		return nil
	})
	pollings.Store(p.ID, p)
	return nil
}

func updatePolling(p *pollingEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := pollings.Load(p.ID); !ok {
		return errInvalidID
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		b.Put([]byte(p.ID), s)
		return nil
	})
	pollings.Store(p.ID, p)
	return nil
}

func deletePolling(pollingID string) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := pollings.Load(pollingID); !ok {
		return errInvalidID
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		b.Delete([]byte(pollingID))
		return nil
	})
	pollings.Delete(pollingID)
	// Delete lines
	for k, v := range lines {
		if v.PollingID1 == pollingID || v.PollingID2 == pollingID {
			deleteLine(k)
		}
	}
	return nil
}

// getNodePollings : ノードを指定してポーリングリストを取得する
func getNodePollings(nodeID string) []pollingEnt {
	ret := []pollingEnt{}
	pollings.Range(func(_, p interface{}) bool {
		if p.(*pollingEnt).NodeID == nodeID {
			ret = append(ret, *p.(*pollingEnt))
		}
		return true
	})
	return ret
}

// getLogPollings : ログを監視するポーリングリストを取得する
func getLogPollings() []pollingEnt {
	ret := []pollingEnt{}
	pollings.Range(func(_, p interface{}) bool {
		t := p.(*pollingEnt).Type
		if t == "syslog" || t == "trap" || t == "netflow" || t == "ipfix" {
			ret = append(ret, *p.(*pollingEnt))
		}
		return true
	})
	return ret
}

func addEventLog(e eventLogEnt) {
	e.Time = time.Now().UnixNano()
	eventLogCh <- e
}

func getEventLogList(startID string, n int) []eventLogEnt {
	ret := []eventLogEnt{}
	if db == nil {
		return ret
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Last(); k != nil && i < n && string(k) != startID; k, v = c.Prev() {
			var e eventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				astilog.Errorf("getEventLogList err=%v", err)
				continue
			}
			ret = append(ret, e)
			i++
		}
		return nil
	})
	return ret
}

func getNodeEventLogs(nodeID string) []eventLogEnt {
	ret := []eventLogEnt{}
	if db == nil {
		return ret
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Last(); k != nil && i < 1000; k, v = c.Prev() {
			var e eventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				astilog.Errorf("getNodeEventLogs err=%v", err)
				continue
			}
			if nodeID != e.NodeID {
				continue
			}
			ret = append(ret, e)
			i++
		}
		return nil
	})
	return ret
}

type logFilterParamEnt struct {
	StartKey    string
	StartTime   int64
	EndTime     int64
	RegexFilter *regexp.Regexp
}

func getFilterParams(filter *filterEnt) *logFilterParamEnt {
	var err error
	var t time.Time
	ret := &logFilterParamEnt{}
	t, err = time.Parse("2006-01-02T15:04 MST", filter.StartTime+" JST")
	if err == nil {
		ret.StartTime = t.UnixNano()
	} else {
		astilog.Errorf("getFilterParams err=%v", err)
		ret.StartTime = time.Now().Add(-time.Hour * 24).UnixNano()
	}
	t, err = time.Parse("2006-01-02T15:04 MST", filter.EndTime+" JST")
	if err == nil {
		ret.EndTime = t.UnixNano()
	} else {
		astilog.Errorf("getFilterParams err=%v", err)
		ret.EndTime = time.Now().UnixNano()
	}
	ret.StartKey = fmt.Sprintf("%016x", ret.StartTime)
	filter.Filter = strings.TrimSpace(filter.Filter)
	if filter.Filter == "" {
		return ret
	}
	ret.RegexFilter, err = regexp.Compile(filter.Filter)
	if err != nil {
		astilog.Errorf("getFilterParams err=%v", err)
		ret.RegexFilter = nil
	}
	return ret
}

func getEventLogs(filter *filterEnt) []eventLogEnt {
	ret := []eventLogEnt{}
	if db == nil {
		return ret
	}
	f := getFilterParams(filter)
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		if b == nil {
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(f.StartKey)); k != nil && i < 100000; k, v = c.Next() {
			var e eventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				astilog.Errorf("getEventLogs err=%v", err)
				continue
			}
			if e.Time < f.StartTime {
				continue
			}
			if e.Time > f.EndTime {
				break
			}
			if f.RegexFilter != nil && !f.RegexFilter.Match(v) {
				continue
			}
			ret = append(ret, e)
			i++
		}
		return nil
	})
	return ret
}

func getLogs(filter *filterEnt) []logEnt {
	ret := []logEnt{}
	if db == nil {
		return ret
	}
	f := getFilterParams(filter)
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(filter.LogType))
		if b == nil {
			astilog.Errorf("getLogs no Bucket=%s", filter.LogType)
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(f.StartKey)); k != nil && i < 100000; k, v = c.Next() {
			var l logEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				astilog.Errorf("getLogs err=%v", err)
				continue
			}
			if l.Time < f.StartTime {
				continue
			}
			if l.Time > f.EndTime {
				break
			}
			if f.RegexFilter != nil && !f.RegexFilter.Match(v) {
				continue
			}
			ret = append(ret, l)
			i++
		}
		return nil
	})
	return ret
}

func addPollingLog(p *pollingEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	pl := pollingLogEnt{
		Time:      time.Now().UnixNano(),
		PollingID: p.ID,
		State:     p.State,
		NumVal:    p.LastVal,
		StrVal:    p.LastResult,
	}
	s, err := json.Marshal(pl)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		b.Put([]byte(makeKey()), s)
		return nil
	})
	return nil
}

func getPollingLog(startTime, endTime, pollingID string) []pollingLogEnt {
	ret := []pollingLogEnt{}
	var st int64
	var et int64
	if t, err := time.Parse("2006-01-02T15:04 MST", startTime+" JST"); err == nil {
		st = t.UnixNano()
	} else {
		astilog.Errorf("getPollingLog err=%v", err)
		st = time.Now().Add(-time.Hour * 24).UnixNano()
	}
	if t, err := time.Parse("2006-01-02T15:04 MST", endTime+" JST"); err == nil {
		et = t.UnixNano()
	} else {
		astilog.Errorf("getFilterParams err=%v", err)
		et = time.Now().UnixNano()
	}
	startKey := fmt.Sprintf("%016x", st)
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			astilog.Errorf("getPollingLog no Bucket getPollingLog")
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(startKey)); k != nil && i < 100000; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			var l pollingLogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				astilog.Errorf("getPollingLog err=%v", err)
				continue
			}
			if l.Time < st {
				continue
			}
			if l.Time > et {
				break
			}
			if l.PollingID != pollingID {
				continue
			}
			ret = append(ret, l)
			i++
		}
		return nil
	})
	return ret
}

func deleteOldLog(bucket string, days int) error {
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days))
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("Bucket %s not found", bucket)
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if st > string(k) {
				break
			}
			c.Delete()
		}
		return nil
	})
}

func deleteOldLogs() {
	if mapConf.LogDays < 1 {
		astilog.Error("mapConf.LogDays < 1 ")
		return
	}
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix"}
	for _, b := range buckets {
		if err := deleteOldLog(b, mapConf.LogDays); err != nil {
			astilog.Errorf("deleteOldLog err=%v")
		}
	}
}

func getMIBModuleList() []string {
	ret := []string{}
	if db == nil {
		return ret
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		if b == nil {
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			ret = append(ret, string(k))
			return nil
		})
		return nil
	})
	return ret
}

func getMIBModule(m string) []byte {
	ret := []byte{}
	if db == nil {
		return ret
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		if b == nil {
			return nil
		}
		ret = b.Get([]byte(m))
		return nil
	})
	return ret
}

func putMIBFileToDB(m, path string) error {
	if db == nil {
		return errDBNotOpen
	}
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		b.Put([]byte(m), d)
		return nil
	})
}

func delMIBModuleFromDB(m string) error {
	if db == nil {
		return errDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("mibdb"))
		b.Delete([]byte(m))
		return nil
	})
}

func closeDB() {
	if db == nil {
		return
	}
	saveLogList([]eventLogEnt{eventLogEnt{
		Type:  "system",
		Level: "info",
		Time:  time.Now().UnixNano(),
		Event: "TWSNMP終了",
	}})
	db.Close()
	db = nil
}

func eventLogger(ctx context.Context) {
	timer1 := time.NewTicker(time.Minute * 5)
	timer2 := time.NewTicker(time.Second * 5)
	list := []eventLogEnt{}
	for {
		select {
		case <-ctx.Done():
			{
				if len(list) > 0 {
					saveLogList(list)
				}
				timer1.Stop()
				timer2.Stop()
				return
			}
		case e := <-eventLogCh:
			{
				list = append(list, e)
				if len(list) > 100 {
					saveLogList(list)
					list = []eventLogEnt{}
				}
			}
		case <-timer1.C:
			{
				deleteOldLogs()
			}
		case <-timer2.C:
			{
				if len(list) > 0 {
					saveLogList(list)
					list = []eventLogEnt{}
				}
			}
		}
	}
}

func saveLogList(list []eventLogEnt) {
	if db == nil {
		return
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		for _, e := range list {
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			err = b.Put([]byte(fmt.Sprintf("%016x", e.Time)), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// bboltに保存する場合のキーを時刻から生成する。
func makeKey() string {
	return fmt.Sprintf("%016x", time.Now().UnixNano())
}
