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
// balloon tree structure against a storage engine.
package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/balloon/storage"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/storage/bolt"
	"github.com/bbva/qed/storage/cache"
)

// Server encapsulates the data and login to start/stop a QED server
type Server struct {
	httpEndpoint string
	dbPath       string
	apiKey       string
	cacheSize    uint64
	storageName  string
	profiling    bool
	tampering    bool

	httpServer      *http.Server
	tamperingServer *http.Server
	profilingServer *http.Server
}

// NewServer synthesizes a new Server based on the parameters it receives.
//
// Note that storageName must be one of 'badger', 'bolt'.
func NewServer(
	httpEndpoint string,
	dbPath string,
	apiKey string,
	cacheSize uint64,
	storageName string,
	profiling bool,
	tampering bool,
) *Server {
	server := new(Server)
	server.httpEndpoint = httpEndpoint
	server.dbPath = dbPath
	server.apiKey = apiKey
	server.cacheSize = cacheSize
	server.storageName = storageName
	server.profiling = profiling
	server.tampering = tampering
	return server
}

func (s *Server) Run() error {

	log.Info("Starting QED server...")

	frozen, leaves, err := buildStorageEngine(s.storageName, s.dbPath)
	balloon, err := buildBalloon(frozen, leaves, s.apiKey, s.cacheSize)
	if err != nil {
		return err
	}

	if s.profiling {
		s.profilingServer = startProfilingServer("localhost:6060")
	}

	if s.tampering {
		s.tamperingServer = startTamperingServer("localhost:8081", leaves.(storage.DeletableStore), hashing.Sha256Hasher)
	}

	s.httpServer = startHTTPServer(s.httpEndpoint, balloon)

	awaitTermSignal(s.Stop)

	log.Info("Stopping server, about to exit...")

	// Give things a few seconds to tidy up
	time.Sleep(time.Second * 2)

	return nil
}

func (s *Server) Stop() {

	if s.tampering {
		log.Info("Tampering enabled: stopping server...")
		if err := s.tamperingServer.Shutdown(nil); err != nil { // TODO include timeout instead nil
			panic(err)
		}
		log.Info("Done.")
	}

	if s.profiling {
		log.Info("Profiling enabled: stopping server...")
		if err := s.profilingServer.Shutdown(nil); err != nil { // TODO include timeout instead nil
			panic(err)
		}
		log.Info("Done.")
	}

	log.Info("Stopping HTTP server...")
	if err := s.httpServer.Shutdown(nil); err != nil { // TODO include timeout instead nil
		panic(err)
	}

	log.Info("Done. Exiting...")
}

func buildStorageEngine(storageName, dbPath string) (storage.Store, storage.Store, error) {
	var frozen, leaves storage.Store
	log.Info("Building storage engine...")
	switch storageName {
	case "badger":
		frozen = badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", dbPath))
		leaves = badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", dbPath))
	case "bolt":
		frozen = bolt.NewBoltStorage(fmt.Sprintf("%s/frozen.db", dbPath), "frozen")
		leaves = bolt.NewBoltStorage(fmt.Sprintf("%s/leaves.db", dbPath), "leaves")
	default:
		log.Error("Please select a valid storage backend")
		return nil, nil, fmt.Errorf("Invalid storage name")
	}
	log.Info("Done.")
	return frozen, leaves, nil
}

func buildBalloon(frozen, leaves storage.Store, apiKey string, cacheSize uint64) (*balloon.HyperBalloon, error) {
	cache := cache.NewSimpleCache(cacheSize)
	hasher := hashing.Sha256Hasher
	history := history.NewTree(frozen, hasher)
	hyper := hyper.NewTree(apiKey, cache, leaves, hasher)
	return balloon.NewHyperBalloon(hasher, history, hyper), nil
}

func startHTTPServer(endpoint string, balloon *balloon.HyperBalloon) *http.Server {

	log.Info("Starting HTTP server...")
	router := apihttp.NewApiHttp(balloon)
	server := &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(router),
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start HTTP Server: %s", err)
		}
	}()
	log.Infof("HTTP server ready and listening on %s", endpoint)

	return server
}

func startProfilingServer(endpoint string) *http.Server {
	log.Info("Starting profiling server...")
	server := &http.Server{
		Addr:    endpoint,
		Handler: nil,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start HTTP Server: %s", err)
		}
	}()

	log.Infof("Profiling server ready and listening on %s", endpoint)

	return server
}

func startTamperingServer(endpoint string, leaves storage.DeletableStore, hasher hashing.Hasher) *http.Server {
	log.Info("Starting tampering server...")
	server := tamperServer(endpoint, leaves.(storage.DeletableStore), hasher)

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start HTTP Server: %s", err)
		}
	}()

	log.Infof("Tampering server ready and listening on %s", endpoint)

	return server
}

func awaitTermSignal(closeFn func()) {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// block main and wait for a signal
	sig := <-signals
	log.Infof("Signal received: %v", sig)

	closeFn()
}

func tamperServer(endpoint string, store storage.DeletableStore, hasher hashing.Hasher) *http.Server {

	type tamperEvent struct {
		Key       []byte
		KeyDigest []byte
		Value     []byte
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

		tp.KeyDigest = hasher(tp.Key)

		switch r.Method {
		case "PATCH":
			get, _ := store.Get(tp.KeyDigest)
			log.Debugf("Get: %v", get)
			log.Debugf("Tamper: %v", store.Add(tp.KeyDigest, tp.Value))

		case "DELETE":
			get, _ := store.Get(tp.KeyDigest)
			log.Debugf("Get: %v", get)
			log.Debugf("Delete: %v", store.Delete(tp.KeyDigest))

		}

		return

	}

	tamperApi := http.NewServeMux()
	tamperApi.HandleFunc("/tamper", apihttp.AuthHandlerMiddleware(http.HandlerFunc(tamper)))

	st := &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(tamperApi),
	}

	return st
}
