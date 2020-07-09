package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
)

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
	router, err := rest.MakeRouter(
		rest.Get("/mapstatus", restAPIGetMapStatus),
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
		Addr:         fmt.Sprintf(":%d", restAPIConf.Port),
		Handler:      api.MakeHandler(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		TLSConfig:    cfg,
	}
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
	High   int
	Low    int
	Warn   int
	Normal int
	Unkown int
	Repair int
	DBSize int64
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
	w.WriteJson(ms)
}
