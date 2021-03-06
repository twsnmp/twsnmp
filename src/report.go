package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/mrichman/godnsbl"
	"github.com/openrdap/rdap"
	"github.com/oschwald/geoip2-golang"
	"go.etcd.io/bbolt"
)

var (
	devices        = make(map[string]*deviceEnt)
	users          = make(map[string]*userEnt)
	flows          = make(map[string]*flowEnt)
	servers        = make(map[string]*serverEnt)
	dennyRules     = make(map[string]bool) // Server:Service:Loc
	allowRules     = make(map[string]*allowRuleEnt)
	deviceReportCh = make(chan *deviceReportEnt, 100)
	userReportCh   = make(chan *userReportEnt, 100)
	flowReportCh   = make(chan *flowReportEnt, 500)
	serviceMap     = make(map[string]string)
	badIPs         = make(map[string]int64)
	protMap        = map[int]string{
		1:   "icmp",
		2:   "igmp",
		6:   "tcp",
		8:   "egp",
		17:  "udp",
		112: "vrrp",
	}
	privateIPBlocks []*net.IPNet
	geoip           *geoip2.Reader
	geoipMap        = make(map[string]string)
)

type deviceReportEnt struct {
	Time int64
	MAC  string
	IP   string
}

type userReportEnt struct {
	Time   int64
	UserID string
	Server string
	Client string
	Ok     bool
}

type flowReportEnt struct {
	Time    int64
	SrcIP   string
	SrcPort int
	DstIP   string
	DstPort int
	Prot    int
	Bytes   int64
}

type deviceEnt struct {
	ID         string // MAC Addr
	Name       string
	IP         string
	Vendor     string
	Services   map[string]int64
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type userEnt struct {
	ID         string // User ID + Server
	UserID     string
	Server     string
	ServerName string
	Clients    map[string]int64
	Total      int
	Ok         int
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type serverEnt struct {
	ID         string //  ID Server
	Server     string
	Services   map[string]int64
	Count      int64
	Bytes      int64
	ServerName string
	Loc        string
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

type flowEnt struct {
	ID         string // ID Client:Server
	Client     string
	Server     string
	Services   map[string]int64
	Count      int64
	Bytes      int64
	ClientName string
	ClientLoc  string
	ServerName string
	ServerLoc  string
	Score      float64
	Penalty    int64
	FirstTime  int64
	LastTime   int64
	UpdateTime int64
}

// allowRuleEnt : 特定のサービスは特定のサーバーに限定するルール
type allowRuleEnt struct {
	Service string // Service
	Servers map[string]bool
}

func initReport() {
	for _, cidr := range []string{
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err == nil {
			privateIPBlocks = append(privateIPBlocks, block)
		}
	}
	openGeoIP()
	_ = loadReport()
}

func openGeoIP() {
	if geoip != nil {
		geoip.Close()
		geoip = nil
	}
	if mapConf.GeoIPPath == "" {
		return
	}
	var err error
	geoip, err = geoip2.Open(mapConf.GeoIPPath)
	if err != nil {
		astiLogger.Errorf("Geoip open err=%v", err)
	}
}

func isPrivateIP(ip net.IP) bool {
	if !ip.IsGlobalUnicast() {
		return true
	}
	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func isGlobalUnicast(ips string) bool {
	ip := net.ParseIP(ips)
	return ip.IsGlobalUnicast()
}

func getLoc(ips string) string {
	if l, ok := geoipMap[ips]; ok {
		return l
	}
	loc := ""
	ip := net.ParseIP(ips)
	if isPrivateIP(ip) {
		loc = "LOCAL,0,0,"
	} else {
		if geoip == nil {
			return loc
		}
		record, err := geoip.City(ip)
		if err == nil {
			loc = fmt.Sprintf("%s,%f,%f,%s", record.Country.IsoCode, record.Location.Latitude, record.Location.Longitude, record.City.Names["en"])
		} else {
			astiLogger.Errorf("getLoc err=%v", err)
			loc = "LOCAL,0,0,"
		}
	}
	geoipMap[ips] = loc
	return loc
}

func loadReport() error {
	if db == nil {
		return errDBNotOpen
	}
	return db.View(func(tx *bbolt.Tx) error {
		r := tx.Bucket([]byte("report"))
		b := r.Bucket([]byte("devices"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var d deviceEnt
				if err := json.Unmarshal(v, &d); err == nil {
					devices[d.ID] = &d
				}
				return nil
			})
		}
		b = r.Bucket([]byte("users"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var u userEnt
				if err := json.Unmarshal(v, &u); err == nil {
					users[u.ID] = &u
				}
				return nil
			})
		}
		b = r.Bucket([]byte("servers"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var s serverEnt
				if err := json.Unmarshal(v, &s); err == nil {
					servers[s.ID] = &s
				}
				return nil
			})
		}
		b = r.Bucket([]byte("flows"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var f flowEnt
				if err := json.Unmarshal(v, &f); err == nil {
					flows[f.ID] = &f
				}
				return nil
			})
		}
		b = r.Bucket([]byte("dennys"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var en bool
				if err := json.Unmarshal(v, &en); err == nil {
					dennyRules[string(k)] = en
				}
				return nil
			})
		}
		b = r.Bucket([]byte("allows"))
		if b != nil {
			_ = b.ForEach(func(k, v []byte) error {
				var as allowRuleEnt
				if err := json.Unmarshal(v, &as); err == nil {
					allowRules[as.Service] = &as
				}
				return nil
			})
		}
		return nil
	})
}

func saveReport(last int64) error {
	if db == nil {
		return errDBNotOpen
	}
	return db.Batch(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		r := b.Bucket([]byte("devices"))
		for _, d := range devices {
			if d.UpdateTime < last {
				continue
			}
			s, err := json.Marshal(d)
			if err != nil {
				return err
			}
			err = r.Put([]byte(d.ID), s)
			if err != nil {
				return err
			}
		}
		r = b.Bucket([]byte("users"))
		for _, u := range users {
			if u.UpdateTime < last {
				continue
			}
			s, err := json.Marshal(u)
			if err != nil {
				return err
			}
			err = r.Put([]byte(u.ID), s)
			if err != nil {
				return err
			}
		}
		r = b.Bucket([]byte("servers"))
		for _, s := range servers {
			if s.UpdateTime < last {
				continue
			}
			js, err := json.Marshal(s)
			if err != nil {
				return err
			}
			err = r.Put([]byte(s.ID), js)
			if err != nil {
				return err
			}
		}
		r = b.Bucket([]byte("flows"))
		for _, f := range flows {
			if f.UpdateTime < last {
				continue
			}
			s, err := json.Marshal(f)
			if err != nil {
				return err
			}
			err = r.Put([]byte(f.ID), s)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func reportBackend(ctx context.Context) {
	initReport()
	timer := time.NewTicker(time.Minute * 5)
	if len(devices) < 1 {
		go checkOldArpLog()
	} else {
		checkOldReport()
		calcScore()
	}
	last := int64(0)
	for {
		select {
		case <-ctx.Done():
			{
				timer.Stop()
				_ = saveReport(0)
				astiLogger.Info("Stop reportBackend")
				return
			}
		case <-timer.C:
			{
				checkOldReport()
				calcScore()
				_ = saveReport(last)
				last = time.Now().UnixNano()
			}
		case r := <-deviceReportCh:
			checkDeviceReport(r)
		case r := <-userReportCh:
			checkUserReport(r)
		case r := <-flowReportCh:
			checkFlowReport(r)
		}
	}
}

func checkUserReport(r *userReportEnt) {
	now := time.Now().UnixNano()
	id := fmt.Sprintf("%s:%s", r.UserID, r.Server)
	u, ok := users[id]
	if ok {
		u.Total++
		if r.Ok {
			u.Ok++
		} else {
			u.Penalty++
		}
		if _, ok := u.Clients[r.Client]; ok {
			u.Clients[r.Client]++
		} else {
			// 複数の場所からログインは問題
			u.Penalty++
			u.Clients[r.Client] = 1
			checkUserClient(u, r.Client)
		}
		u.LastTime = r.Time
		u.UpdateTime = now
		return
	}
	u = &userEnt{
		ID:         id,
		UserID:     r.UserID,
		Server:     r.Server,
		ServerName: findNameFromIP(r.Server),
		Clients:    make(map[string]int64),
		Total:      1,
		FirstTime:  r.Time,
		LastTime:   r.Time,
		UpdateTime: now,
	}
	u.Clients[r.Client] = 1
	checkUserClient(u, r.Client)
	if r.Ok {
		u.Ok = 1
	} else {
		u.Penalty = 1
	}
	users[id] = u
	astiLogger.Debugf("add users %v", u)
}

func checkUserClient(u *userEnt, client string) {
	if !strings.Contains(client, ".") {
		return
	}
	loc := getLoc(client)
	a := strings.Split(loc, ",")
	if len(a) > 0 {
		loc = a[0]
	}
	// DNSで解決できない場合
	if client == findNameFromIP(client) {
		u.Penalty++
	}
	if loc != "" && loc != "LOCAL" {
		id := fmt.Sprintf("*:*:%s", loc)
		if _, ok := dennyRules[id]; ok {
			u.Penalty++
		}
	}
	if u.Penalty > 0 {
		if _, ok := badIPs[client]; !ok {
			badIPs[client] = u.Penalty
		}
	}
}

// getFlowDir : クライアント、サーバー、サービスを決定するアルゴリズム
func getFlowDir(r *flowReportEnt) (server, client, service string) {
	guc1 := isGlobalUnicast(r.SrcIP)
	guc2 := isGlobalUnicast(r.DstIP)
	if !guc1 && !guc2 {
		// 両方ユニキャストでない場合は含めない。
		return
	}
	s1, ok1 := getServiceName(r.Prot, r.SrcPort)
	s2, ok2 := getServiceName(r.Prot, r.DstPort)
	if ok1 {
		if ok2 {
			if r.SrcPort < r.DstPort || !guc1 {
				// ポート番号の小さい方を優先、または、マルチキャストはサーバーとする
				server = r.SrcIP
				client = r.DstIP
				service = s1
			} else if r.SrcPort == r.DstPort {
				if _, ok := flows[fmt.Sprintf("%s:%s", r.DstIP, r.SrcIP)]; ok || !guc2 {
					// 既に登録済みか、マルチキャストをサーバーとする
					server = r.DstIP
					client = r.SrcIP
					service = s2
				} else {
					server = r.SrcIP
					client = r.DstIP
					service = s1
				}
			} else {
				server = r.DstIP
				client = r.SrcIP
				service = s2
			}
		} else {
			server = r.SrcIP
			client = r.DstIP
			service = s1
		}
	} else {
		if ok2 {
			server = r.DstIP
			client = r.SrcIP
			service = s2
		} else {
			if r.SrcPort < r.DstPort || !guc1 {
				server = r.SrcIP
				client = r.DstIP
				service = s1
			} else {
				server = r.DstIP
				client = r.SrcIP
				service = s2
			}
		}
	}
	return
}

func checkFlowReport(r *flowReportEnt) {
	server, client, service := getFlowDir(r)
	if server == "" {
		astiLogger.Warnf("Skip flow report %v", r)
		return
	}
	checkServerReport(server, service, r.Bytes, r.Time)
	now := time.Now().UnixNano()
	id := fmt.Sprintf("%s:%s", client, server)
	f, ok := flows[id]
	if ok {
		if _, ok := f.Services[service]; ok {
			f.Services[service]++
		} else {
			f.Services[service] = 1
			setFlowPenalty(f)
		}
		if f.ServerLoc == "" {
			f.ServerLoc = getLoc(f.Server)
		}
		if f.ClientLoc == "" {
			f.ClientLoc = getLoc(f.Client)
		}
		f.Bytes += r.Bytes
		f.Count++
		f.LastTime = r.Time
		f.UpdateTime = now
		return
	}
	f = &flowEnt{
		ID:         id,
		Client:     client,
		Server:     server,
		Services:   make(map[string]int64),
		Count:      1,
		Bytes:      r.Bytes,
		ServerLoc:  getLoc(server),
		ClientLoc:  getLoc(client),
		ServerName: findNameFromIP(server),
		ClientName: findNameFromIP(client),
		FirstTime:  r.Time,
		LastTime:   r.Time,
		UpdateTime: now,
	}
	f.Services[service] = 1
	setFlowPenalty(f)
	flows[id] = f
}

func checkServerReport(server, service string, bytes, t int64) {
	if !strings.Contains(service, "/") {
		return
	}
	now := time.Now().UnixNano()
	id := server
	s, ok := servers[id]
	if ok {
		if _, ok := s.Services[service]; ok {
			s.Services[service]++
		} else {
			s.Services[service] = 1
			setServerPenalty(s)
		}
		s.Count++
		s.Bytes += bytes
		s.LastTime = t
		s.UpdateTime = now
		return
	}
	s = &serverEnt{
		ID:         id,
		Server:     server,
		Services:   make(map[string]int64),
		ServerName: findNameFromIP(server),
		Loc:        getLoc(server),
		Count:      1,
		Bytes:      bytes,
		FirstTime:  t,
		LastTime:   t,
		UpdateTime: now,
	}
	s.Services[service] = 1
	setServerPenalty(s)
	servers[id] = s
	astiLogger.Debugf("add server %v", s)
}

func setFlowPenalty(f *flowEnt) {
	loc := ""
	if f.ServerLoc != "" {
		a := strings.Split(f.ServerLoc, ",")
		if len(a) > 0 {
			loc = a[0]
		}
	}
	f.Penalty = 0
	ids := []string{}
	for service := range f.Services {
		ids = append(ids, fmt.Sprintf("*:%s:*", service))
		if loc != "" {
			ids = append(ids, fmt.Sprintf("*:%s:%s", service, loc))
		}
		if as, ok := allowRules[service]; ok {
			if e, ok := as.Servers[f.Server]; !ok {
				if e {
					f.Penalty++
				}
			}
		}
	}
	ids = append(ids, fmt.Sprintf("%s:*:*", f.Server))
	if loc != "" {
		ids = append(ids, fmt.Sprintf("*:*:%s", loc))
	}
	for _, id := range ids {
		if e, ok := dennyRules[id]; ok {
			if e {
				f.Penalty++
			}
		}
	}
	// DNSで解決できない場合
	if f.ServerName == f.Server {
		f.Penalty++
	}
	if f.Penalty > 0 {
		if n, ok := badIPs[f.Client]; !ok || n < f.Penalty {
			badIPs[f.Client] = f.Penalty
		}
	}
}

func setServerPenalty(s *serverEnt) {
	loc := ""
	if s.Loc != "" {
		a := strings.Split(s.Loc, ",")
		if len(a) > 0 {
			loc = a[0]
		}
	}
	s.Penalty = 0
	ids := []string{}
	for service := range s.Services {
		ids = append(ids, fmt.Sprintf("*:%s:*", service))
		if loc != "" {
			ids = append(ids, fmt.Sprintf("*:%s:%s", service, loc))
		}
		if as, ok := allowRules[service]; ok {
			if e, ok := as.Servers[s.Server]; !ok {
				if e {
					s.Penalty++
				}
			}
		}
	}
	if loc != "" {
		ids = append(ids, fmt.Sprintf("*:*:%s", loc))
	}
	for _, id := range ids {
		if e, ok := dennyRules[id]; ok {
			if e {
				s.Penalty++
			}
		}
	}
	// DNSで解決できない場合
	if s.ServerName == s.Server {
		s.Penalty++
	}
}

func getServiceName(prot, port int) (string, bool) {
	if p, ok := protMap[prot]; ok {
		k := fmt.Sprintf("%d/%s", port, p)
		if s, ok := serviceMap[k]; ok {
			return s, true
		}
		return p, false
	}
	return fmt.Sprintf("prot(%d)", prot), false
}

func checkDeviceReport(r *deviceReportEnt) {
	ip := r.IP
	mac := r.MAC
	d, ok := devices[mac]
	if ok {
		if d.IP != ip {
			d.IP = ip
			d.Name = findNameFromIP(ip)
			setDevicePenalty(d)
			// IPアドレスが変わるもの
			d.Penalty++
		}
		d.LastTime = r.Time
		d.UpdateTime = time.Now().UnixNano()
		return
	}
	d = &deviceEnt{
		ID:         mac,
		IP:         ip,
		Name:       findNameFromIP(ip),
		Vendor:     oui.Find(mac),
		FirstTime:  r.Time,
		LastTime:   r.Time,
		UpdateTime: time.Now().UnixNano(),
	}
	setDevicePenalty(d)
	devices[mac] = d
	astiLogger.Debugf("add devices %v", d)
}

func setDevicePenalty(d *deviceEnt) {
	// ベンダー禁止のもの
	if d.Vendor == "Unknown" {
		d.Penalty++
	}
	// ホスト名が不明なもの
	if d.IP == d.Name {
		d.Penalty++
	}
	ip := net.ParseIP(d.IP)
	if !isPrivateIP(ip) {
		d.Penalty++
	}
}

func checkOldArpLog() {
	if db == nil {
		return
	}
	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("arplog"))
		if b == nil {
			astiLogger.Error("checkOldArpLog no arplog bucket")
			return nil
		}
		_ = b.ForEach(func(k, v []byte) error {
			var l logEnt
			if err := json.Unmarshal(v, &l); err == nil {
				a := strings.Split(l.Log, ",")
				if len(a) > 2 {
					if strings.HasPrefix(a[2], "FF") || strings.HasPrefix(a[2], "01") {
						return nil
					}
					deviceReportCh <- &deviceReportEnt{
						IP:   a[1],
						MAC:  a[2],
						Time: l.Time,
					}
				}
			}
			return nil
		})
		return nil
	})
}

func findNameFromIP(ip string) string {
	if names, err := net.LookupAddr(ip); err == nil && len(names) > 0 {
		return names[0]
	}
	for _, n := range nodes {
		if n.IP == ip {
			return n.Name
		}
	}
	return ip
}

func checkOldReport() {
	oh := -24
	if len(servers) > 10000 {
		oh = -12 / (len(servers) / 10000)
		if oh > -3 {
			oh = -3
		}
	}
	old := time.Now().Add(time.Hour * time.Duration(oh)).UnixNano()
	tooOld := time.Now().AddDate(0, 0, -mapConf.LogDays).UnixNano()
	checkOldServers(old, tooOld)
	checkOldFlows(old, tooOld)
	checkOldDevices(old)
	checkOldUsers(old)
}

func checkOldServers(old, tooOld int64) {
	count := 0
	ids := []string{}
	for _, s := range servers {
		if s.LastTime < old {
			if s.LastTime < tooOld || s.LastTime-s.FirstTime < 3600*1000*1000*1000 {
				ids = append(ids, s.ID)
			} else {
				for k, n := range s.Services {
					if n < 10 {
						delete(s.Services, k)
					}
				}
				if len(s.Services) < 1 {
					ids = append(ids, s.ID)
				}
			}
		}
	}
	for _, id := range ids {
		_ = deleteReport("servers", id)
		count++
	}
	if count > 0 {
		astiLogger.Infof("DeleteSevers=%d", count)
	}
}

func checkOldFlows(old, tooOld int64) {
	count := 0
	ids := []string{}
	for _, f := range flows {
		if f.LastTime < old {
			if f.LastTime < tooOld || f.LastTime-f.FirstTime < 3600*1000*1000*1000 {
				ids = append(ids, f.ID)
			} else {
				for k, n := range f.Services {
					if n < 10 {
						delete(f.Services, k)
					}
				}
				if len(f.Services) < 1 {
					ids = append(ids, f.ID)
				}
			}
		}
	}
	for _, id := range ids {
		_ = deleteReport("flows", id)
		count++
	}
	if count > 0 {
		astiLogger.Infof("DeleteFlows=%d", count)
	}
}

func checkOldDevices(tooOld int64) {
	count := 0
	ids := []string{}
	for _, d := range devices {
		if d.LastTime < tooOld {
			ids = append(ids, d.ID)
		}
	}
	for _, id := range ids {
		_ = deleteReport("devices", id)
		count++
	}
	if count > 0 {
		astiLogger.Infof("DeleteDevices=%d", count)
	}
}

func checkOldUsers(tooOld int64) {
	count := 0
	ids := []string{}
	for _, u := range users {
		if u.LastTime < tooOld {
			ids = append(ids, u.ID)
		}
	}
	for _, id := range ids {
		_ = deleteReport("users", id)
		count++
	}
	if count > 0 {
		astiLogger.Infof("DeleteUsers=%d", count)
	}
}

func calcScore() {
	calcDeviceScore()
	calcServerScore()
	calcFlowScore()
	calcUserScore()
	badIPs = make(map[string]int64)
}

func calcDeviceScore() {
	var xs []float64
	for _, d := range devices {
		if n, ok := badIPs[d.IP]; ok {
			d.Penalty += n
		}
		if d.Penalty > 100 {
			d.Penalty = 100
		}
		xs = append(xs, float64(100-d.Penalty))
	}
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	for _, d := range devices {
		d.Score = ((10 * (float64(100-d.Penalty) - m) / sd) + 50)
	}
}

func calcFlowScore() {
	var xs []float64
	for _, f := range flows {
		if f.Penalty > 100 {
			f.Penalty = 100
		}
		xs = append(xs, float64(100-f.Penalty))
	}
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	for _, f := range flows {
		f.Score = ((10 * (float64(100-f.Penalty) - m) / sd) + 50)
	}
}

func calcUserScore() {
	var xs []float64
	for _, u := range users {
		if u.Penalty > 100 {
			u.Penalty = 100
		}
		xs = append(xs, float64(100-u.Penalty))
	}
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	for _, u := range users {
		u.Score = ((10 * (float64(100-u.Penalty) - m) / sd) + 50)
	}
}

func calcServerScore() {
	var xs []float64
	for _, s := range servers {
		if s.Penalty > 100 {
			s.Penalty = 100
		}
		xs = append(xs, float64(100-s.Penalty))
	}
	m, sd := getMeanSD(&xs)
	if sd == 0 {
		return
	}
	for _, s := range servers {
		s.Score = ((10 * (float64(100-s.Penalty) - m) / sd) + 50)
	}
}

func getMeanSD(xs *[]float64) (float64, float64) {
	m, err := stats.Mean(*xs)
	if err != nil {
		return 0, 0
	}
	sd, err := stats.StandardDeviation(*xs)
	if err != nil {
		return 0, 0
	}
	return m, sd
}

func deleteReport(report, id string) error {
	if db == nil {
		return errDBNotOpen
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte(report))
			if r != nil {
				_ = r.Delete([]byte(id))
			}
		}
		return nil
	})
	if report == "devices" {
		delete(devices, id)
	} else if report == "users" {
		delete(users, id)
	} else if report == "servers" {
		delete(servers, id)
	} else if report == "flows" {
		delete(flows, id)
	}
	return nil
}

func resetPenalty(report string) {
	if report == "devices" {
		for _, d := range devices {
			d.Penalty = 0
			setDevicePenalty(d)
			d.UpdateTime = time.Now().UnixNano()
		}
		calcDeviceScore()
	} else if report == "users" {
		for _, u := range users {
			u.Penalty = 0
			u.UpdateTime = time.Now().UnixNano()
		}
		calcUserScore()
	} else if report == "servers" {
		for _, s := range servers {
			if s.Loc == "" {
				s.Loc = getLoc(s.Server)
			}
			setServerPenalty(s)
			s.UpdateTime = time.Now().UnixNano()
		}
		calcServerScore()
	} else if report == "flows" {
		for _, f := range flows {
			if f.ServerLoc == "" {
				f.ServerLoc = getLoc(f.Server)
			}
			if f.ClientLoc == "" {
				f.ClientLoc = getLoc(f.Client)
			}
			setFlowPenalty(f)
			f.UpdateTime = time.Now().UnixNano()
		}
		calcFlowScore()
	}
}

func clearAllReport() error {
	if db == nil {
		return errDBNotOpen
	}
	_ = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			for _, r := range []string{"devices", "flows", "users", "servers"} {
				_ = b.DeleteBucket([]byte(r))
				_, _ = b.CreateBucketIfNotExists([]byte(r))
			}
		}
		return nil
	})
	devices = make(map[string]*deviceEnt)
	users = make(map[string]*userEnt)
	flows = make(map[string]*flowEnt)
	servers = make(map[string]*serverEnt)
	return nil
}

func loadServiceMap(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		l := strings.TrimSpace(s.Text())
		if len(l) < 1 || strings.HasPrefix(l, "#") {
			continue
		}
		f := strings.Fields(l)
		if len(f) < 2 {
			continue
		}
		sn := f[0]
		a := strings.Split(f[1], "/")
		if len(a) > 1 {
			sn += "/" + a[1]
		}
		serviceMap[f[1]] = sn
	}
	return nil
}

func addAllowRule(service, server string) error {
	if db == nil {
		return errDBNotOpen
	}
	as, ok := allowRules[service]
	if ok {
		as.Servers[server] = true
	} else {
		as = &allowRuleEnt{
			Service: service,
			Servers: map[string]bool{server: true},
		}
		allowRules[service] = as
	}
	js, err := json.Marshal(as)
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("allows"))
			if r != nil {
				_ = r.Put([]byte(service), js)
			}
		}
		return nil
	})
}

func deleteAllowRule(id string) error {
	if db == nil {
		return errDBNotOpen
	}
	a := strings.Split(id, ":")
	if len(a) != 2 {
		return fmt.Errorf("deleteAllowRule bad id %s", id)
	}
	server := a[0]
	service := a[1]
	as, ok := allowRules[service]
	if !ok {
		return nil
	}
	delete(as.Servers, server)
	js := []byte{}
	if len(as.Servers) > 0 {
		js, _ = json.Marshal(as)
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("allows"))
			if r != nil {
				if len(js) < 1 {
					_ = r.Delete([]byte(service))
				} else {
					_ = r.Put([]byte(service), js)
				}
			}
		}
		return nil
	})
}

func addDennyRule(id string) error {
	if db == nil {
		return errDBNotOpen
	}
	dennyRules[id] = true
	js, err := json.Marshal(dennyRules[id])
	if err != nil {
		return err
	}
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("dennys"))
			if r != nil {
				_ = r.Put([]byte(id), js)
			}
		}
		return nil
	})
}

func deleteDennyRule(id string) error {
	if db == nil {
		return errDBNotOpen
	}
	_, ok := dennyRules[id]
	if !ok {
		return nil
	}
	delete(dennyRules, id)
	return db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("report"))
		if b != nil {
			r := b.Bucket([]byte("dennys"))
			if r != nil {
				_ = r.Delete([]byte(id))
			}
		}
		return nil
	})
}

type ipInfoCache struct {
	Time   int64
	IPInfo *[][]string
}

var ipInfoCacheMap = make(map[string]*ipInfoCache)

var blacklists = []string{
	"b.barracudacentral.org",
	"bl.spamcop.net",
	"blacklist.woody.ch",
	"bogons.cymru.com",
	"cbl.abuseat.org",
	"combined.abuse.ch",
	"db.wpbl.info",
	"dnsbl-1.uceprotect.net",
	"dnsbl-2.uceprotect.net",
	"dnsbl-3.uceprotect.net",
	"dnsbl.dronebl.org",
	"dnsbl.inps.de",
	"dnsbl.sorbs.net",
	"drone.abuse.ch",
	"duinv.aupads.org",
	"dul.dnsbl.sorbs.net",
	"dyna.spamrats.com",
	"dynip.rothen.com",
	"http.dnsbl.sorbs.net",
	"ips.backscatterer.org",
	"ix.dnsbl.manitu.net",
	"korea.services.net",
	"misc.dnsbl.sorbs.net",
	"noptr.spamrats.com",
	"orvedb.aupads.org",
	"pbl.spamhaus.org",
	"proxy.bl.gweep.ca",
	"psbl.surriel.com",
	"relays.bl.gweep.ca",
	"relays.nether.net",
	"sbl.spamhaus.org",
	"smtp.dnsbl.sorbs.net",
	"socks.dnsbl.sorbs.net",
	"spam.abuse.ch",
	"spam.dnsbl.sorbs.net",
	"spam.spamrats.com",
	"spamrbl.imp.ch",
	"ubl.unsubscore.com",
	"virus.rbl.jp",
	"web.dnsbl.sorbs.net",
	"wormrbl.imp.ch",
	"xbl.spamhaus.org",
	"zen.spamhaus.org",
	"zombie.dnsbl.sorbs.net",

	"z.mailspike.net",
	"spamsources.fabel.dk",
	"spambot.bls.digibase.ca",
	"spam.dnsbl.anonmails.de",
	"singular.ttk.pte.hu",
	"all.s5h.net",
	"ubl.lashback.com",
	"dnsbl.spfbl.net",
}

func getIPInfo(ip string) *[][]string {
	if c, ok := ipInfoCacheMap[ip]; ok {
		if c.Time > time.Now().Unix()-60*60*24*7 {
			return c.IPInfo
		}
	}
	ret := [][]string{}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		client := &rdap.Client{}
		ri, err := client.QueryIP(ip)
		if err != nil {
			astiLogger.Errorf("RDAP QueryIP error=%v", err)
			return
		}
		ret = append(ret, []string{"RDAP:IP Version", ri.IPVersion}) //IPバージョン
		ret = append(ret, []string{"RDAP:Type", ri.Type})            // 種類
		ret = append(ret, []string{"RDAP:Handole", ri.Handle})       //範囲
		ret = append(ret, []string{"RDAP:Name", ri.Name})            // 所有者
		ret = append(ret, []string{"RDAP:Country", ri.Country})      // 国
		ret = append(ret, []string{"RDAP:Whois Server", ri.Port43})  // Whoisの情報源
	}()
	rblMap := &sync.Map{}
	for i, source := range blacklists {
		wg.Add(1)
		go func(i int, source string) {
			defer wg.Done()
			rbl := godnsbl.Lookup(source, ip)
			if len(rbl.Results) > 0 && rbl.Results[0].Listed {
				rblMap.Store(source, `<i class="fas fa-exclamation-circle state state_high"></i>Listed :`+rbl.Results[0].Text)
			} else {
				rblMap.Store(source, `<i class="fas fa-check-circle state state_repair"></i>Not Listed`)
			}
		}(i, source)
	}
	wg.Wait()
	rblMap.Range(func(key, value interface{}) bool {
		ret = append(ret, []string{"DNSBL:" + key.(string), value.(string)})
		return true
	})
	ipInfoCacheMap[ip] = &ipInfoCache{
		Time:   time.Now().Unix(),
		IPInfo: &ret,
	}
	return &ret
}
