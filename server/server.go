/*
    Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/

// Package server implements the server initialization for the api.apihttp and
// ballon tree structure against a storage engine.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/balloon/storage"
	"github.com/bbva/qed/balloon/storage/badger"
	"github.com/bbva/qed/balloon/storage/bolt"
	"github.com/bbva/qed/balloon/storage/cache"
	"github.com/bbva/qed/log"
)

func NewServer(
	httpEndpoint string,
	dbPath string,
	apiKey string,
	cacheSize uint64,
	storageName string,
	profiling bool,
	tampering bool,
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

	if profiling {
		// start profiler
		go func() {
			log.Info(http.ListenAndServe("localhost:6060", nil))
		}()
	}

	if tampering {
		log.Debug("tampering")
		go tamperServer(leaves.(storage.DeletableStore))
	}

	router := apihttp.NewApiHttp(balloon)

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

func tamperServer(store storage.DeletableStore) {

	type tamperEvent struct {
		Key   []byte
		Value []byte
	}

	tamper := func(w http.ResponseWriter, r *http.Request) {

		// Make sure we can only be called with an HTTP POST request.
		if !(r.Method == "PATCH" || r.Method == "DELETE") {
			w.Header().Set("Allow", "PATCH, DELETE")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}

		var tp tamperEvent
		err := json.NewDecoder(r.Body).Decode(&tp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "PATCH":
			get, _ := store.Get(tp.Key)
			log.Debugf("Get: %v", get)
			log.Debugf("Tamper: %v", store.Add(tp.Key, tp.Value))

		case "DELETE":
			get, _ := store.Get(tp.Key)
			log.Debugf("Get: %v", get)
			log.Debugf("Delete: %v", store.Delete(tp.Key))

		}

		return

	}

	tamperApi := http.NewServeMux()
	tamperApi.HandleFunc("/tamper", apihttp.AuthHandlerMiddleware(http.HandlerFunc(tamper)))

	st := &http.Server{
		Addr:    "localhost:8081",
		Handler: logHandler(tamperApi),
	}

	log.Error(st.ListenAndServe())

}
