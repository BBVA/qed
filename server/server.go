// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package server

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"qed/api/apihttp"
	"qed/balloon"
	"qed/balloon/hashing"
	"qed/balloon/history"
	"qed/balloon/hyper"
	"qed/balloon/storage"
	"qed/balloon/storage/badger"
	"qed/balloon/storage/bolt"
	"qed/balloon/storage/cache"
	"qed/log"
)

func NewServer(
	httpEndpoint string,
	dbPath string,
	apiKey string,
	cacheSize uint64,
	storageName string,
	tamperable bool,
) *http.Server {

	var frozen, leaves storage.Store

	switch storageName {
	case "badger":
		frozen = badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", dbPath))
		leaves = badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", dbPath))
	case "bolt":
		frozen = bolt.NewBoltStorage(fmt.Sprintf("%s/frozen.db", dbPath), "frozen")
		leaves = bolt.NewBoltStorage(fmt.Sprintf("%s/leaves.db", dbPath), "leaves")
	default:
		log.Error("Please select a valid storage backend")
	}

	cache := cache.NewSimpleCache(cacheSize)
	hasher := hashing.Sha256Hasher
	history := history.NewTree(frozen, hasher)
	hyper := hyper.NewTree(apiKey, cache, leaves, hasher)
	balloon := balloon.NewHyperBalloon(hasher, history, hyper)

	// start profiler
	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	tamperOpts := apiHttp.TamperOpts{
		tamperable,
		leaves
	}

	router := apihttp.NewApiHttp(balloon, tamperOpts)

	return &http.Server{
		Addr:    httpEndpoint,
		Handler: logHandler(router),
	}

}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

// WriteLog Logs the Http Status for a request into fileHandler and returns a
// httphandler function which is a wrapper to log the requests.
func logHandler(handle http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		start := time.Now()
		writer := statusWriter{w, 0, 0}
		handle.ServeHTTP(&writer, request)
		latency := time.Now().Sub(start)

		log.Debugf("Request: lat %d %+v", latency, request)
		if writer.status >= 400 {
			log.Infof("Bad Request: %d %+v", latency, request)
		}
	}
}
