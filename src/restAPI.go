package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	api.Use(&rest.AccessLogApacheMiddleware{
		Format: rest.CombinedLogFormat,
	})
	api.Use(&rest.TimerMiddleware{})
	api.Use(&rest.RecorderMiddleware{})
	api.Use(&rest.RecoverMiddleware{})
	api.Use(&rest.GzipMiddleware{})
	api.Use(&rest.ContentTypeCheckerMiddleware{})
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
		MinVersion:               tls.VersionTLS12,
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
	http.Handle("/api/", twsnmpWebHandler(http.StripPrefix("/api", api.MakeHandler())))
	http.Handle("/", twsnmpWebHandler(http.StripPrefix("/", http.FileServer(http.Dir(resAppPath)))))
	go func(r *http.Server) {
		if err := r.ListenAndServeTLS("", ""); err != nil {
			astiLogger.Errorf("restAPI err=%v", err)
		}
		r.Close()
	}(restAPI)
	astiLogger.Infof("start restAPI")
	return nil
}

func twsnmpWebHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Frame-Options", "SAMEORIGIN")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Add("Pragma", "no-cache")
		w.Header().Add("Cache-Control", "no-cache, no-store, must-revalidate")
		if strings.HasSuffix(r.URL.Path, ".js") ||
			strings.HasSuffix(r.URL.Path, ".css") ||
			strings.HasSuffix(r.URL.Path, ".ico") ||
			strings.HasSuffix(r.URL.Path, ".woff") ||
			strings.HasSuffix(r.URL.Path, ".woff2") ||
			strings.HasSuffix(r.URL.Path, ".ttf") ||
			strings.HasSuffix(r.URL.Path, ".png") ||
			strings.HasSuffix(r.URL.Path, ".svg") ||
			r.URL.Path == "/" ||
			strings.HasSuffix(r.URL.Path, "mapstatus") ||
			strings.HasSuffix(r.URL.Path, "mapdata") ||
			strings.HasSuffix(r.URL.Path, "index.html") {
			h.ServeHTTP(w, r)
			return
		}
		astiLogger.Infof("Not Found %v", r.URL.Path)
		http.NotFound(w, r)
	})
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
	Unknown   int
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
			ms.Unknown++
		}
	}
	if ms.High > 0 {
		ms.State = "high"
	} else if ms.Low > 0 {
		ms.State = "low"
	} else if ms.Warn > 0 {
		ms.State = "warn"
	} else if ms.Normal+ms.Repair > 0 {
		ms.State = "normal"
	} else {
		ms.State = "unknown"
	}
	ms.DBSize = dbStats.NSize
	ms.DBSizeStr = dbStats.Size
	_ = w.WriteJson(ms)
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
	_ = w.WriteJson(&restAPIMapData)
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
	var body string
	var err error
	for i := 0; !ok && i <= p.Retry; i++ {
		startTime := time.Now().UnixNano()
		body, err = doTWSNMPGet(n, p)
		endTime := time.Now().UnixNano()
		if err != nil {
			continue
		}
		rTime = endTime - startTime
		ok = true
	}
	p.LastVal = float64(rTime)
	if ok {
		var ms restMapStatusEnt
		if err := json.Unmarshal([]byte(body), &ms); err != nil {
			setPollingError("twsnmp", p, err)
			return
		}
		lr := make(map[string]string)
		lr["rtt"] = fmt.Sprintf("%f", p.LastVal)
		lr["state"] = ms.State
		lr["high"] = fmt.Sprintf("%d", ms.High)
		lr["low"] = fmt.Sprintf("%d", ms.Low)
		lr["warn"] = fmt.Sprintf("%d", ms.Warn)
		lr["normal"] = fmt.Sprintf("%d", ms.Normal)
		lr["repair"] = fmt.Sprintf("%d", ms.Repair)
		lr["dbsize"] = fmt.Sprintf("%d", ms.DBSize)
		p.LastResult = makeLastResult(lr)
		setPollingState(p, ms.State)
		return
	}
	setPollingError("twsnmp", p, err)
}

func doTWSNMPGet(n *nodeEnt, p *pollingEnt) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.Timeout)*time.Second)
	defer cancel()
	url := fmt.Sprintf("https://%s:8192/api/mapstatus", n.IP)
	if n.URL != "" {
		url = n.URL + "/api/mapstatus"
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(n.User, n.Password)
	resp, err := insecureClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
