package main

import (
	"encoding/json"
	"fmt"
	"time"
	"strings"
	"context"
	"regexp"
	astilog "github.com/asticode/go-astilog"
	"go.etcd.io/bbolt"
)

var (
	db *bbolt.DB
	// Data on Memory
	mapConf      mapConfEnt
	discoverConf discoverConfEnt
	nodes        = make(map[string]*nodeEnt)
	lines        = make(map[string]*lineEnt)
	pollings     = make(map[string]*pollingEnt)
	eventLogCh   = make(chan eventLogEnt,100)
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
	ID      string
	NodeID1 string
	PollingID1 string
	State1     string
	NodeID2 string
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

type logEnt struct {
	Time   int64 // UnixNano()
	Type   string
	Log    string
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
	EnableSyslogd   bool
	EnableTrapd     bool
	EnableNetflowd  bool
	BackImg        string
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
	buckets := []string{"config", "nodes", "lines", "pollings", "logs", "pollingLogs","syslog","trap","netflow","ipfix"}
	mapConf.Community = "public"
	mapConf.PollInt = 60
	mapConf.Retry = 1
	mapConf.Timeout = 1
	mapConf.LogDispSize = 5000
	mapConf.LogDays = 14
	discoverConf.Community = "public"
	discoverConf.Retry = 1
	discoverConf.Timeout = 1
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
		return nil
	})
	if err == nil && bSaveConf {
		saveMapConfToDB()
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
				nodes[n.ID] =  &n
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
					pollings[p.ID] = &p
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
		if _,ok := nodes[n.ID]; !ok {
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
	nodes[n.ID]  = n
	return nil
}

func updateNode(n *nodeEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	if _,ok := nodes[n.ID]; !ok {
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
	if _,ok := nodes[nodeID]; !ok {
		return errInvalidID
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		b.Delete([]byte(nodeID))
		return nil
	})
	delete(nodes,nodeID)
	for k,v := range pollings {
		if v.NodeID == nodeID {
			deletePolling(k)
		}
	}
	return nil
}

func findNodeFromIP(ip string) *nodeEnt {
	for _,n := range nodes {
		if n.IP == ip {
			return n
		} 
	}
	return nil
}

func addLine(l *lineEnt) error {
	for {
		l.ID = makeKey()
		if _,ok := lines[l.ID]; !ok {
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
	if _,ok := lines[l.ID];!ok {
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
	if _,ok := lines[lineID]; !ok {
		return errInvalidID
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("lines"))
		b.Delete([]byte(lineID))
		return nil
	})
	delete(lines,lineID)
	return nil
}

// pollings
func addPolling(p *pollingEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	for {
		p.ID = makeKey()
		if _,ok := pollings[p.ID];!ok {
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
	pollings[p.ID] = p
	return nil
}

func updatePolling(p *pollingEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	if _,ok := pollings[p.ID]; !ok {
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
	pollings[p.ID] = p
	return nil
}

func deletePolling(pollingID string) error {
	if db == nil {
		return errDBNotOpen
	}
	if _,ok := pollings[pollingID]; !ok {
		return errInvalidID
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollings"))
		b.Delete([]byte(pollingID))
		return nil
	})
	delete(pollings,pollingID)
	// Delete lines
	for k,v := range lines {
		if v.PollingID1 == pollingID || v.PollingID2 == pollingID{
			deleteLine(k)
		}
	}
	return nil
}

// getNodePollings : ノードを指定してポーリングリストを取得する
func getNodePollings(nodeID string) []pollingEnt {
	ret := []pollingEnt{}
	for _, p := range pollings {
		if p.NodeID == nodeID {
			ret = append(ret, *p)
		}
	}
	return ret
}

// getLogPollings : ログを監視するポーリングリストを取得する
func getLogPollings() []pollingEnt {
	ret := []pollingEnt{}
	for _, p := range pollings {
		if p.Type == "syslog" || p.Type == "trap" ||  p.Type == "netflow" ||  p.Type == "ipfix" {
			ret = append(ret, *p)
		}
	}
	return ret
}


func addEventLog(e eventLogEnt){
	e.Time = time.Now().UnixNano()
	eventLogCh <- e
}

func getEventLogList(startID string,n int) []eventLogEnt{
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
		for k,v := c.Last(); k != nil && i < n && string(k) != startID; k,v = c.Prev(){
			var e eventLogEnt
			err := json.Unmarshal(v,&e)
			if err != nil {
				astilog.Errorf("getEventLogList err=%v",err)
				continue
			}
			ret = append(ret,e)
			i++
		}
		return nil
	})
	return ret
}

func getNodeEventLogs(nodeID string) []eventLogEnt{
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
		for k,v := c.Last(); k != nil && i < 100000; k,v = c.Prev(){
			var e eventLogEnt
			err := json.Unmarshal(v,&e)
			if err != nil {
				astilog.Errorf("getNodeEventLogs err=%v",err)
				continue
			}
			if nodeID != e.NodeID {
				continue
			}
			ret = append(ret,e)
			i++
		}
		return nil
	})
	return ret
}

type logFilterParamEnt struct {
	StartKey  string
	StartTime int64
	EndTime   int64
	RegexFilter    *regexp.Regexp
}

func getFilterParams(filter *filterEnt) *logFilterParamEnt{
	var err error
	var t time.Time
	ret  := &logFilterParamEnt{} 
	t,err = time.Parse("2006-01-02T15:04 MST",filter.StartTime + " JST")
	if err == nil {
		ret.StartTime = t.UnixNano()
	} else {
		astilog.Errorf("getFilterParams err=%v",err)
		ret.StartTime = time.Now().Add(-time.Hour*24).UnixNano()
	}
	t ,err = time.Parse("2006-01-02T15:04 MST",filter.EndTime + " JST")
	if err == nil {
		ret.EndTime = t.UnixNano()
	} else {
		astilog.Errorf("getFilterParams err=%v",err)
		ret.EndTime = time.Now().UnixNano()
	}
	ret.StartKey = fmt.Sprintf("%016x",ret.StartTime)
	filter.Filter = strings.TrimSpace(filter.Filter)
	if filter.Filter == "" {
		return ret
	}
	ret.RegexFilter,err = regexp.Compile(filter.Filter)
	if err != nil {
		astilog.Errorf("getFilterParams err=%v",err)
		ret.RegexFilter = nil
	}
	return ret
}

func getEventLogs(filter *filterEnt) []eventLogEnt{
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
		for k,v := c.Seek([]byte(f.StartKey)); k != nil && i < 100000; k,v = c.Next(){
			var e eventLogEnt
			err := json.Unmarshal(v,&e)
			if err != nil {
				astilog.Errorf("getEventLogs err=%v",err)
				continue
			}
			if e.Time < f.StartTime {
				continue
			}
			if e.Time > f.EndTime {
				break
			}
			if f.RegexFilter != nil  && !f.RegexFilter.Match(v){
					continue
			}
			ret = append(ret,e)
			i++
		}
		return nil
	})
	return ret
}

func getLogs(filter *filterEnt) []logEnt{
	ret := []logEnt{}
	if db == nil {
		return ret
	}
	f := getFilterParams(filter)
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(filter.LogType))
		if b == nil {
			astilog.Errorf("getLogs no Bucket=%s",filter.LogType)
			return nil
		}
		c := b.Cursor()
		i := 0
		for k,v := c.Seek([]byte(f.StartKey)); k != nil && i < 100000; k,v = c.Next(){
			var l logEnt
			err := json.Unmarshal(v,&l)
			if err != nil {
				astilog.Errorf("getLogs err=%v",err)
				continue
			}
			if l.Time < f.StartTime {
				continue
			}
			if l.Time > f.EndTime {
				break
			}
			if f.RegexFilter != nil  && !f.RegexFilter.Match(v){
					continue
			}
			ret = append(ret,l)
			i++
		}
		return nil
	})
	return ret
}

func deleteOldLog(bucket string,days int) error {
	st := fmt.Sprintf("%016x",time.Now().AddDate(0,0,-days))
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("Bucket %s not found",bucket)
		}
		c := b.Cursor()
		for k,_ := c.First(); k != nil ; k,_ = c.Next(){
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
	buckets := []string{"logs", "pollingLogs","syslog","trap","netflow","ipfix"}
	for _,b := range buckets {
		if err := deleteOldLog(b,mapConf.LogDays); err != nil {
			astilog.Errorf("deleteOldLog err=%v")
		}
	}
}

func closeDB() {
	if db == nil {
		return
	}
	saveLogList([]eventLogEnt{eventLogEnt{
		Type:"system",
		Level:"info",
		Time: time.Now().UnixNano(),
		Event: "TWSNMP終了",
	}})
	db.Close()
	db = nil
}

func eventLogger(ctx context.Context) {
	list := []eventLogEnt{}
	for {
		select {
			case <- ctx.Done():{
				if len(list) > 0 {
					saveLogList(list)
				}
				return
			}
			case e := <- eventLogCh:{
				list = append(list,e)
				if len(list) > 1000 {
					saveLogList(list)
					list =[]eventLogEnt{}
				}
			}
			case <- time.Tick(time.Minute * 5):{
				deleteOldLogs()
			}
			case <- time.Tick(time.Second * 10):{
				if len(list) > 0 {
					saveLogList(list)
					list =[]eventLogEnt{}
				}
			}
		}
	}
}

func saveLogList(list []eventLogEnt){
	if db == nil {
		return
	}
	db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("logs"))
		for _,e := range list {
			s, err := json.Marshal(e)
			if err != nil {
				return err
			}
			err = b.Put([]byte(fmt.Sprintf("%016x",e.Time)), s)
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
