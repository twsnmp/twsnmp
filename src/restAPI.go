package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
)

var resAppPath string
var restAPI *http.Server
var muRestAPI sync.Mutex

func setupRestAPI() error {
	if err := stopRestAPI(); err != nil {
		astiLogger.Errorf("restAPI err=%v", err)
	}
	muRestAPI.Lock()
	defer muRestAPI.Unlock()
	if restAPIConf.Port < 1024 {
		return nil
	}
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	api.Use(&rest.AuthBasicMiddleware{
		Realm: "TWSNMP API",
		Authenticator: func(userId string, password string) bool {
			if userId == restAPIConf.User && password == restAPIConf.Password {
				return true
			}
			return false
		},
	})
	router, err := rest.MakeRouter(
		rest.Get("/mapstatus", restAPIGetMapStatus),
		rest.Get("/mapdata", restAPIGetMapData),
	)
	if err != nil {
		astiLogger.Errorf("restAPI err=%v", err)
		return err
	}
	keyPem := getRawKeyPem(mapConf.PrivateKey)
	cert, err := tls.X509KeyPair([]byte(mapConf.TLSCert), []byte(keyPem))
	if err != nil {
		astiLogger.Errorf("restAPI err=%v", err)
		return err
	}
	api.SetApp(router)
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// CipherSuites: []uint16{
		// 	tls.TLS_AES_128_GCM_SHA256,
		// 	tls.TLS_AES_256_GCM_SHA384,
		// },
		// MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       true,
	}
	restAPI = &http.Server{
		Addr: fmt.Sprintf(":%d", restAPIConf.Port),
		//		Handler:      api.MakeHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    cfg,
	}
	http.Handle("/api/", http.StripPrefix("/api", api.MakeHandler()))
	http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(resAppPath))))
	go func(r *http.Server) {
		if err := r.ListenAndServeTLS("", ""); err != nil {
			astiLogger.Errorf("restAPI err=%v", err)
		}
		r.Close()
	}(restAPI)
	astiLogger.Infof("start restAPI")
	return nil
}

func stopRestAPI() error {
	muRestAPI.Lock()
	defer muRestAPI.Unlock()
	if restAPI == nil {
		return nil
	}
	if err := restAPI.Shutdown(context.Background()); err != nil {
		restAPI = nil
		return err
	}
	restAPI = nil
	return nil
}

// API
type restMapStatusEnt struct {
	High      int
	Low       int
	Warn      int
	Normal    int
	Repair    int
	Unkown    int
	DBSize    int64
	DBSizeStr string
	State     string
}

func restAPIGetMapStatus(w rest.ResponseWriter, req *rest.Request) {
	ms := &restMapStatusEnt{}
	for _, n := range nodes {
		switch n.State {
		case "high":
			ms.High++
		case "low":
			ms.Low++
		case "warn":
			ms.Warn++
		case "normal":
			ms.Normal++
		case "repair":
			ms.Repair++
		default:
			ms.Unkown++
		}
	}
	if ms.High > 0 {
		ms.State = "high"
	} else if ms.Low > 0 {
		ms.State = "low"
	} else if ms.Normal > 0 {
		ms.State = "normal"
	} else {
		ms.State = "unknown"
	}
	ms.DBSize = dbStats.NSize
	ms.DBSizeStr = dbStats.Size
	w.WriteJson(ms)
}

type restAPIMapDataEnt struct {
	LastTime int64
	MapName  string
	BackImg  bool
	Nodes    map[string]*restAPINodeEnt
	Lines    map[string]*lineEnt
	Pollings []*pollingEnt
	Logs     []eventLogEnt
}

type restAPINodeEnt struct {
	ID    string
	Name  string
	Descr string
	Icon  string
	State string
	X     int
	Y     int
	IP    string
	MAC   string
}

var restAPIMapData = restAPIMapDataEnt{
	Nodes: make(map[string]*restAPINodeEnt),
	Lines: make(map[string]*lineEnt),
}

func makeRestAPIMapData() {
	if restAPIMapData.LastTime > time.Now().Unix()-60 {
		return
	}
	restAPIMapData.MapName = mapConf.MapName
	restAPIMapData.LastTime = time.Now().Unix()
	restAPIMapData.BackImg = mapConf.BackImg != ""
	for id, n := range nodes {
		restAPIMapData.Nodes[id] = &restAPINodeEnt{
			ID:    id,
			Name:  n.Name,
			Descr: n.Descr,
			Icon:  n.Icon,
			State: n.State,
			X:     n.X,
			Y:     n.Y,
			IP:    n.IP,
			MAC:   n.MAC,
		}
	}
	for id, l := range lines {
		restAPIMapData.Lines[id] = l
	}
	pollings.Range(func(_, v interface{}) bool {
		p := v.(*pollingEnt)
		restAPIMapData.Pollings = append(restAPIMapData.Pollings, p)
		return true
	})
	restAPIMapData.Logs = getEventLogList(fmt.Sprintf("%016x", time.Now().Unix()-3600*24), 1000)
}

func restAPIGetMapData(w rest.ResponseWriter, req *rest.Request) {
	makeRestAPIMapData()
	w.WriteJson(&restAPIMapData)
}

// TWSNMPへのポーリング
func doPollingTWSNMP(p *pollingEnt) {
	n, ok := nodes[p.NodeID]
	if !ok {
		setPollingError("twsnmp", p, fmt.Errorf("Node not found"))
		return
	}
	ok = false
	var rTime int64
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		err := doTWSNMPGet(n, p)
		endTime := time.Now().UnixNano()
		if err != nil {
			setPollingError("twsnmp", p, err)
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		var ms restMapStatusEnt
		if err := json.Unmarshal([]byte(p.LastResult), &ms); err == nil {
			setPollingState(p, ms.State)
		} else {
			setPollingState(p, "unknown")
		}
	} else {
		setPollingState(p, "unknown")
	}
}

func doTWSNMPGet(n *nodeEnt, p *pollingEnt) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Timeout)*time.Second)
	defer cancel()
	url := fmt.Sprintf("https://%s:8192/api/mapstatus", n.IP)
	if n.URL != "" {
		url = n.URL + "/api/mapstatus"
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(n.User, n.Password)
	resp, err := insecureClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	p.LastResult = string(b)
	return nil
}
