package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"

	"go.etcd.io/bbolt"
)

var (
	db *bbolt.DB
	// Data on Memory
	mapConf           mapConfEnt
	notifyConf        notifyConfEnt
	discoverConf      discoverConfEnt
	prevDBStats       bbolt.Stats
	dbStats           dbStatsEnt
	dbOpenTime        time.Time
	nodes             = make(map[string]*nodeEnt)
	lines             = make(map[string]*lineEnt)
	pollings          = sync.Map{}
	eventLogCh        = make(chan eventLogEnt, 100)
	stopEventLoggerCh = make(chan bool)
	mainWindowInfo    windowInfoEnt
	pollingTemplates  = make(map[string]*pollingTemplateEnt)
)

const (
	// MaxDispLog : ログの検索結果の最大値
	MaxDispLog = 20000
	// MaxDelLog : ログ削除処理の最大削除数
	MaxDelLog = 500000
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
	MAC       string
	SnmpMode  string
	Community string
	User      string
	Password  string
	PublicKey string
	URL       string
	Type      string
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
	NextTime   int64
	LastTime   int64
	LastResult string
	LastVal    float64
	State      string
}

type pollingTemplateEnt struct {
	ID       string
	Name     string
	Type     string
	Polling  string
	Level    string
	NodeType string
	Descr    string
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
	SnmpMode       string
	Community      string
	User           string
	Password       string
	PublicKey      string
	PrivateKey     string
	TLSCert        string
	EnableSyslogd  bool
	EnableTrapd    bool
	EnableNetflowd bool
	BackImg        string
	GeoIPPath      string
	GrokPath       string
	ArpWatchLevel  string
	AILevel        string
	AIThreshold    int
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
	ExecCmd            string
}

type discoverConfEnt struct {
	SnmpMode string
	StartIP  string
	EndIP    string
	Timeout  int
	Retry    int
	X        int
	Y        int
}

type aiResult struct {
	PollingID string
	LastTime  int64
	LossData  [][]float64
	ScoreData [][]float64
}

type dbStatsEnt struct {
	Time             string
	Size             string
	TotalWrite       string
	LastWrite        string
	PeakWrite        string
	AvgWrite         string
	StartTime        string
	Speed            string
	Peak             string
	Rate             float64
	BackupConfigOnly bool
	BackupDaily      bool
	BackupFile       string
	BackupTime       string
}

type windowInfoEnt struct {
	Top    int
	Left   int
	Width  int
	Height int
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
	prevDBStats = db.Stats()
	dbOpenTime = time.Now()
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
	loadPollingTemplateFromDB()
	return nil
}

func initDB() error {
	buckets := []string{"config", "nodes", "lines", "pollings", "logs", "pollingLogs",
		"syslog", "trap", "netflow", "ipfix", "arplog", "mibdb", "arp", "ai", "report", "pollingTemplates"}
	reports := []string{"devices", "users", "flows", "servers", "allows", "dennys"}
	mapConf.Community = "public"
	mapConf.PollInt = 60
	mapConf.Retry = 1
	mapConf.Timeout = 1
	mapConf.LogDispSize = 5000
	mapConf.LogDays = 14
	mapConf.ArpWatchLevel = "info"
	mapConf.AILevel = "info"
	mapConf.AIThreshold = 81
	mapConf.Community = "public"
	discoverConf.Retry = 1
	discoverConf.Timeout = 1
	notifyConf.InsecureSkipVerify = true
	notifyConf.Interval = 60
	notifyConf.Subject = "TWSNMPからの通知"
	notifyConf.Level = "none"
	mainWindowInfo.Width = 1024
	mainWindowInfo.Height = 800
	return db.Update(func(tx *bbolt.Tx) error {
		for _, b := range buckets {
			pb, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
			if b == "report" {
				for _, r := range reports {
					if _, err := pb.CreateBucketIfNotExists([]byte(r)); err != nil {
						return err
					}
				}
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
			astiLogger.Error(fmt.Sprintf("Unmarshal mapConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("discoverConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &discoverConf); err != nil {
			astiLogger.Error(fmt.Sprintf("Unmarshal discoverConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("notifyConf"))
		if v == nil {
			return nil
		}
		if err := json.Unmarshal(v, &notifyConf); err != nil {
			astiLogger.Error(fmt.Sprintf("Unmarshal notifyConf from DB error=%v", err))
			return err
		}
		v = b.Get([]byte("mainWindowInfo"))
		if v != nil {
			if err := json.Unmarshal(v, &mainWindowInfo); err != nil {
				astiLogger.Error(fmt.Sprintf("Unmarshal mainWindowInfo from DB error=%v", err))
			}
		}
		var p dbBackupParamEnt
		v = b.Get([]byte("backup"))
		if v != nil {
			if err := json.Unmarshal(v, &p); err != nil {
				astiLogger.Error(fmt.Sprintf("Unmarshal mainWinbackupdowInfo from DB error=%v", err))
			} else {
				if p.BackupFile != "" && p.Daily {
					dbStats.BackupConfigOnly = p.ConfigOnly
					dbStats.BackupFile = p.BackupFile
					dbStats.BackupDaily = p.Daily
					now := time.Now()
					d := 0
					if now.Hour() > 2 {
						d = 1
					}
					nextBackup = time.Date(now.Year(), now.Month(), now.Day()+d, 3, 0, 0, 0, time.Local).UnixNano()
				}
			}
		}
		return nil
	})
	if err == nil && mapConf.PrivateKey == "" {
		initSecurityKey()
	}
	if mainWindowInfo.Width < 100 || mainWindowInfo.Height < 100 {
		mainWindowInfo.Width = 1024
		mainWindowInfo.Height = 800
		mainWindowInfo.Top = -1
	}
	if err == nil && bSaveConf {
		saveMapConfToDB()
		saveNotifyConfToDB()
		saveDiscoverConfToDB()
		saveMainWindowInfoToDB()
	}
	return err
}

func initSecurityKey() {
	key, err := genPrivateKey(4096)
	if err != nil {
		astiLogger.Errorf("initSecurityKey err=%v", err)
		return
	}
	pubkey, err := getSSHPublicKey(key)
	if err != nil {
		astiLogger.Errorf("initSecurityKey err=%v", err)
		return
	}
	cert, err := genSelfSignCert(key)
	if err != nil {
		astiLogger.Errorf("initSecurityKey err=%v", err)
		return
	}
	mapConf.PrivateKey = key
	mapConf.PublicKey = pubkey
	mapConf.TLSCert = cert
	astiLogger.Infof("initSecurityKey Public Key=%v", pubkey)
	saveMapConfToDB()
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

func saveBackupParamToDB(p *dbBackupParamEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	s, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("Bucket config is nil")
		}
		b.Put([]byte("backup"), s)
		return nil
	})
}

func saveMainWindowInfoToDB() error {
	if db == nil {
		return errDBNotOpen
	}
	s, err := json.Marshal(mainWindowInfo)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return fmt.Errorf("Bucket config is nil")
		}
		b.Put([]byte("mainWindowInfo"), s)
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
	p.LastTime = time.Now().UnixNano()
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
	clearPollingLog(pollingID)
	deleteAIReesult(pollingID)
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

// getPollings : ポーリングリストを取得する
func getPollings() []pollingEnt {
	ret := []pollingEnt{}
	pollings.Range(func(_, p interface{}) bool {
		ret = append(ret, *p.(*pollingEnt))
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
				astiLogger.Errorf("getEventLogList err=%v", err)
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
				astiLogger.Errorf("getNodeEventLogs err=%v", err)
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

func parseFilter(f string) string {
	f = strings.TrimSpace(f)
	if f == "``" {
		return ""
	}
	if strings.HasPrefix(f, "`") && strings.HasSuffix(f, "`") {
		return f[1 : len(f)-1]
	}
	f = regexp.QuoteMeta(f)
	f = strings.ReplaceAll(f, "\\*", ".+")
	return f
}

func getFilterParams(filter *filterEnt) *logFilterParamEnt {
	var err error
	var t time.Time
	ret := &logFilterParamEnt{}
	t, err = time.Parse("2006-01-02T15:04 MST", filter.StartTime+" JST")
	if err == nil {
		ret.StartTime = t.UnixNano()
	} else {
		astiLogger.Errorf("getFilterParams err=%v", err)
		ret.StartTime = time.Now().Add(-time.Hour * 24).UnixNano()
	}
	t, err = time.Parse("2006-01-02T15:04 MST", filter.EndTime+" JST")
	if err == nil {
		ret.EndTime = t.UnixNano()
	} else {
		astiLogger.Errorf("getFilterParams err=%v", err)
		ret.EndTime = time.Now().UnixNano()
	}
	ret.StartKey = fmt.Sprintf("%016x", ret.StartTime)
	filter.Filter = strings.TrimSpace(filter.Filter)
	if filter.Filter == "" {
		return ret
	}
	fs := parseFilter(filter.Filter)
	ret.RegexFilter, err = regexp.Compile(fs)
	if err != nil {
		astiLogger.Errorf("getFilterParams err=%v", err)
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
		for k, v := c.Seek([]byte(f.StartKey)); k != nil && i < MaxDispLog; k, v = c.Next() {
			var e eventLogEnt
			err := json.Unmarshal(v, &e)
			if err != nil {
				astiLogger.Errorf("getEventLogs err=%v", err)
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
			astiLogger.Errorf("getLogs no Bucket=%s", filter.LogType)
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(f.StartKey)); k != nil && i < MaxDispLog; k, v = c.Next() {
			if bytes.HasSuffix(v, []byte{0, 0, 255, 255}) {
				v = deCompressLog(v)
			}
			var l logEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				astiLogger.Errorf("getLogs err=%v", err)
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
		astiLogger.Errorf("getPollingLog err=%v", err)
		st = time.Now().Add(-time.Hour * 24).UnixNano()
	}
	if t, err := time.Parse("2006-01-02T15:04 MST", endTime+" JST"); err == nil {
		et = t.UnixNano()
	} else {
		astiLogger.Errorf("getFilterParams err=%v", err)
		et = time.Now().UnixNano()
	}
	startKey := fmt.Sprintf("%016x", st)
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			astiLogger.Errorf("getPollingLog no Bucket getPollingLog")
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.Seek([]byte(startKey)); k != nil && i < MaxDispLog; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			var l pollingLogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				astiLogger.Errorf("getPollingLog err=%v", err)
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

func getAllPollingLog(pollingID string) []pollingLogEnt {
	ret := []pollingLogEnt{}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			astiLogger.Errorf("getPollingLog no Bucket getPollingLog")
			return nil
		}
		c := b.Cursor()
		i := 0
		for k, v := c.First(); k != nil && i < MaxDispLog*100; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			var l pollingLogEnt
			err := json.Unmarshal(v, &l)
			if err != nil {
				astiLogger.Errorf("getPollingLog err=%v", err)
				continue
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

func clearPollingLog(pollingID string) error {
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingLogs"))
		if b == nil {
			return fmt.Errorf("Bucket pollingLogs not found")
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if !bytes.Contains(v, []byte(pollingID)) {
				continue
			}
			c.Delete()
		}
		b = tx.Bucket([]byte("ai"))
		if b != nil {
			b.Delete([]byte(pollingID))
		}
		return nil
	})
}

var delCount int

func deleteOldLog(bucket string, days int) error {
	st := fmt.Sprintf("%016x", time.Now().AddDate(0, 0, -days).UnixNano())
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("Bucket %s not found", bucket)
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if st < string(k) || delCount > MaxDelLog {
				break
			}
			c.Delete()
			delCount++
		}
		return nil
	})
}

func deleteOldLogs() {
	delCount = 0
	if mapConf.LogDays < 1 {
		astiLogger.Error("mapConf.LogDays < 1 ")
		return
	}
	buckets := []string{"logs", "pollingLogs", "syslog", "trap", "netflow", "ipfix", "arplog"}
	for _, b := range buckets {
		if err := deleteOldLog(b, mapConf.LogDays); err != nil {
			astiLogger.Errorf("deleteOldLog err=%v", err)
		}
	}
	if delCount > 0 {
		astiLogger.Infof("DeleteLogs=%d", delCount)
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
	timer1 := time.NewTicker(time.Minute * 2)
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

var peakPS float64
var peakWrite int

func updateDBStats() {
	if db == nil {
		return
	}
	s := db.Stats()
	d := s.Sub(&prevDBStats)
	var dbSize int64
	db.View(func(tx *bbolt.Tx) error {
		dbSize = tx.Size()
		return nil
	})
	dbStats.Size = humanize.Bytes(uint64(dbSize))
	dbStats.TotalWrite = humanize.Comma(int64(s.TxStats.Write))
	dbStats.LastWrite = humanize.Comma(int64(d.TxStats.Write))
	if peakWrite < d.TxStats.Write {
		peakWrite = d.TxStats.Write
		dbStats.PeakWrite = dbStats.LastWrite
	}
	// 初回は計算しない。
	if peakWrite > 0 && dbStats.Time != "" {
		dbStats.Rate = 100 * float64(d.TxStats.Write) / float64(peakWrite)
		dbStats.StartTime = humanize.Time(dbOpenTime)
		dbot := time.Now().Sub(dbOpenTime).Seconds()
		if dbot > 0 {
			dbStats.AvgWrite = humanize.SI(float64(s.TxStats.Write)/dbot, "Write/Sec")
		}
	}
	dt := d.TxStats.WriteTime.Seconds()
	if dt != 0 {
		ps := float64(d.TxStats.Write) / dt
		dbStats.Speed = humanize.SI(ps, "Write/Sec")
		if peakPS < ps {
			peakPS = ps
			dbStats.Peak = dbStats.Speed
		}
	} else {
		dbStats.Speed = "Unkown"
	}
	dbStats.Time = time.Now().Format("15:04:05")
	prevDBStats = s

	if dbStats.BackupFile != "" && nextBackup != 0 && nextBackup < time.Now().UnixNano() {
		nextBackup += (24 * 3600 * 1000 * 1000 * 1000)
		go func() {
			astiLogger.Infof("Backup start = %s", dbStats.BackupFile)
			addEventLog(eventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ開始:" + dbStats.BackupFile,
			})
			if err := backupDB(); err != nil {
				astiLogger.Errorf("backupDB err=%v", err)
			}
			astiLogger.Infof("Backup end = %s", dbStats.BackupFile)
			addEventLog(eventLogEnt{
				Type:  "system",
				Level: "info",
				Event: "バックアップ終了:" + dbStats.BackupFile,
			})
		}()
		dbStats.BackupTime = dbStats.Time
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

func loadArpTableFromDB() error {
	if db == nil {
		return errDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			arpTable[string(k)] = string(v)
			return nil
		})
		return nil
	})
}

func updateArpEnt(ip, mac string) error {
	arpTable[ip] = mac
	if db == nil {
		return errDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		b.Put([]byte(ip), []byte(mac))
		return nil
	})
}

func resetArpTable() error {
	arpTable = make(map[string]string)
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arp"))
		if b == nil {
			return fmt.Errorf("Bucket arp not found")
		}
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			c.Delete()
		}
		return nil
	})
}

func saveAIResultToDB(res *aiResult) error {
	if db == nil {
		return errDBNotOpen
	}
	s, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return fmt.Errorf("Bucket ai is nil")
		}
		b.Put([]byte(res.PollingID), s)
		return nil
	})
}

func loadAIReesult(id string) (*aiResult, error) {
	var ret aiResult
	r := ""
	if db == nil {
		return &ret, errDBNotOpen
	}
	db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		if b == nil {
			return nil
		}
		tmp := b.Get([]byte(id))
		if tmp != nil {
			r = string(tmp)
		}
		return nil
	})
	if r == "" {
		return &ret, nil
	}
	if err := json.Unmarshal([]byte(r), &ret); err != nil {
		return &ret, err
	}
	return &ret, nil
}

func deleteAIReesult(id string) error {
	if db == nil {
		return errDBNotOpen
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("ai"))
		b.Delete([]byte(id))
		return nil
	})
}

var stopBackup = false
var nextBackup int64
var dbBackupSize int64
var dstDB *bbolt.DB
var dstTx *bbolt.Tx

func backupDB() error {
	if db == nil {
		return errDBNotOpen
	}
	if dstDB != nil {
		return fmt.Errorf("Backup in progress")
	}
	os.Remove(dbStats.BackupFile)
	var err error
	dstDB, err = bbolt.Open(dbStats.BackupFile, 0600, nil)
	if err != nil {
		return err
	}
	defer func() {
		dstDB.Close()
		dstDB = nil
	}()
	dstTx, err = dstDB.Begin(true)
	if err != nil {
		return err
	}
	defer dstTx.Rollback()
	err = db.View(func(srcTx *bbolt.Tx) error {
		return srcTx.ForEach(func(name []byte, b *bbolt.Bucket) error {
			return walkBucket(b, nil, name, nil, b.Sequence())
		})
	})
	if err != nil {
		return err
	}
	if !dbStats.BackupConfigOnly {
		mapConfTmp := mapConf
		mapConfTmp.EnableNetflowd = false
		mapConfTmp.EnableSyslogd = false
		mapConfTmp.EnableTrapd = false
		mapConfTmp.LogDays = 0
		if s, err := json.Marshal(mapConfTmp); err == nil {
			if b := dstTx.Bucket([]byte("config")); b != nil {
				b.Put([]byte("mapConf"), s)
			}
		}
	}
	dstTx.Commit()
	return nil
}

var configBuckets = []string{"config", "nodes", "lines", "pollings", "mibdb"}

func walkBucket(b *bbolt.Bucket, keypath [][]byte, k, v []byte, seq uint64) error {
	if stopBackup {
		return fmt.Errorf("Stop Backup")
	}
	if dbStats.BackupConfigOnly && v == nil {
		c := false
		for _, cbn := range configBuckets {
			if k != nil && cbn == string(k) {
				c = true
				break
			}
		}
		if !c {
			return nil
		}
	}
	if dbBackupSize > 64*1024 {
		dstTx.Commit()
		var err error
		dstTx, err = dstDB.Begin(true)
		if err != nil {
			return err
		}
		dbBackupSize = 0
	}
	// Execute callback.
	if err := walkFunc(keypath, k, v, seq); err != nil {
		return err
	}
	dbBackupSize += int64(len(k) + len(v))

	// If this is not a bucket then stop.
	if v != nil {
		return nil
	}

	// Iterate over each child key/value.
	keypath = append(keypath, k)
	return b.ForEach(func(k, v []byte) error {
		if v == nil {
			bkt := b.Bucket(k)
			return walkBucket(bkt, keypath, k, nil, bkt.Sequence())
		}
		return walkBucket(b, keypath, k, v, b.Sequence())
	})
}

func walkFunc(keys [][]byte, k, v []byte, seq uint64) error {
	// Create bucket on the root transaction if this is the first level.
	nk := len(keys)
	if nk == 0 {
		bkt, err := dstTx.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := bkt.SetSequence(seq); err != nil {
			return err
		}
		return nil
	}
	// Create buckets on subsequent levels, if necessary.
	b := dstTx.Bucket(keys[0])
	if nk > 1 {
		for _, k := range keys[1:] {
			b = b.Bucket(k)
		}
	}
	// Fill the entire page for best compaction.
	b.FillPercent = 1.0
	// If there is no value then this is a bucket call.
	if v == nil {
		bkt, err := b.CreateBucket(k)
		if err != nil {
			return err
		}
		if err := bkt.SetSequence(seq); err != nil {
			return err
		}
		return nil
	}
	// Otherwise treat it as a key/value pair.
	return b.Put(k, v)
}

func loadPollingTemplateFromDB() error {
	if db == nil {
		return errDBNotOpen
	}
	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		if b == nil {
			return nil
		}
		b.ForEach(func(k, v []byte) error {
			var pt pollingTemplateEnt
			if err := json.Unmarshal(v, &pt); err == nil {
				pollingTemplates[pt.ID] = &pt
			}
			return nil
		})
		return nil
	})
	return err
}

func getSha1Key(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func addPollingTemplate(pt *pollingTemplateEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	pt.ID = getSha1Key(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if _, ok := pollingTemplates[pt.ID]; ok {
		return fmt.Errorf("duplicate template")
	}
	s, err := json.Marshal(pt)
	if err != nil {
		return err
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		b.Put([]byte(pt.ID), s)
		return nil
	})
	pollingTemplates[pt.ID] = pt
	return nil
}

func updatePollingTemplate(pt *pollingTemplateEnt) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := pollingTemplates[pt.ID]; !ok {
		return errInvalidID
	}
	// 更新後に同じ内容のテンプレートがないか確認する
	newID := getSha1Key(pt.Name + ":" + pt.Type + ":" + pt.NodeType + ":" + pt.Polling)
	if _, ok := pollingTemplates[newID]; ok {
		return fmt.Errorf("duplicate template")
	}
	// なければ、削除してから追加する
	deletePollingTemplate(pt.ID)
	pt.ID = newID
	return addPollingTemplate(pt)
}

func deletePollingTemplate(id string) error {
	if db == nil {
		return errDBNotOpen
	}
	if _, ok := pollingTemplates[id]; !ok {
		return errInvalidID
	}
	db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("pollingTemplates"))
		b.Delete([]byte(id))
		return nil
	})
	delete(pollingTemplates, id)
	return nil
}
