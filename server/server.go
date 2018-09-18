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
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" // this wil enable the default profiling capabilities
	"os"
	"os/signal"
	"syscall"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/api/tampering"
	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/balloon/common"
	"github.com/bbva/qed/balloon/history"
	"github.com/bbva/qed/balloon/hyper"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/sign"
	"github.com/bbva/qed/storage/badger"
)

type Store interface {
	Add(key []byte, value []byte) error
	GetRange(start, end []byte) [][]byte
	Get(key []byte) ([]byte, error)
	Close() error
}

// Server encapsulates the data and login to start/stop a QED server
type Server struct {
	httpEndpoint string
	dbPath       string
	apiKey       string
	cacheSize    uint64
	storages     []Store

	httpServer      *http.Server
	tamperingServer *http.Server
	profilingServer *http.Server
}

// NewServer synthesizes a new Server based on the parameters it receives.
// Note that storageName must be one of 'badger'.
func NewServer(
	httpEndpoint string,
	dbPath string,
	apiKey string,
	cacheSize uint64,
	storageName string,
	profiling bool,
	tamper bool,
	signer sign.Signer,
) *Server {

	storages := make([]Store, 0, 0)

	frozen, leaves, err := buildStorageEngine(storageName, dbPath)
	storages = append(storages, frozen, leaves)
	balloon, err := buildBalloon(frozen, leaves, apiKey, cacheSize)
	if err != nil {
		log.Error(err)
	}

	server := &Server{
		httpEndpoint,
		dbPath,
		apiKey,
		cacheSize,
		storages,
		newHTTPServer(httpEndpoint, balloon, signer),
		nil,
		nil,
	}
	if tamper {
		server.tamperingServer = newTamperingServer("localhost:8081", leaves.(tampering.DeletableStore), hashing.NewSha256Hasher())
	}

	if profiling {
		server.profilingServer = newProfilingServer("localhost:6060")
	}

	return server

}

// Run will start the server in a non-blockable fashion.
func (s *Server) Run() error {

	log.Debugf("Starting QED server...")

	if s.profilingServer != nil {
		go func() {
			log.Debugf("	* Starting profiling HTTP server in addr: localhost:6060")
			if err := s.profilingServer.ListenAndServe(); err != http.ErrServerClosed {
				log.Errorf("Can't start profiling HTTP server: %s", err)
			}
		}()
	}

	if s.tamperingServer != nil {
		go func() {
			log.Debug("	* Starting tampering HTTP server in addr: localhost:8081")
			if err := s.tamperingServer.ListenAndServe(); err != http.ErrServerClosed {
				log.Errorf("Can't start tampering HTTP server: %s", err)
			}
		}()
	}

	go func() {
		log.Debug("	* Starting QED API HTTP server in addr: ", s.httpEndpoint)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start QED API HTTP Server: %s", err)
		}
	}()

	log.Debugf(" ready on %s\n", s.httpEndpoint)

	awaitTermSignal(s.Stop)

	log.Debug("Stopping server, about to exit...")

	return nil
}

// Stop will close all the channels from the mux servers.
func (s *Server) Stop() {

	if s.tamperingServer != nil {
		log.Debugf("Tampering enabled: stopping server...")
		if err := s.tamperingServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
			log.Error(err)
		}
		log.Debugf("Done.\n")
	}

	if s.profilingServer != nil {
		log.Debugf("Profiling enabled: stopping server...")
		if err := s.profilingServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
			log.Error(err)
		}
		log.Debugf("Done.\n")
	}

	log.Debugf("Stopping HTTP server...")
	if err := s.httpServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		log.Error(err)
	}

	for _, st := range s.storages {
		st.Close()
	}

	log.Debugf("Done. Exiting...\n")
}

func buildStorageEngine(storageName, dbPath string) (Store, Store, error) {
	var frozen, leaves Store
	log.Debugf("Building storage engine...")
	switch storageName {
	case "badger":
		frozen = badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", dbPath))
		leaves = badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", dbPath))
	default:
		log.Error("Please select a valid storage backend")
		return nil, nil, fmt.Errorf("Invalid storage name")
	}
	log.Debug("Done.")
	return frozen, leaves, nil
}

func buildBalloon(frozen, leaves Store, apiKey string, cacheSize uint64) (*balloon.HyperBalloon, error) {
	cache := common.NewSimpleCache(cacheSize)
	history := history.NewTree(apiKey, frozen, hashing.NewSha256Hasher())
	hyper := hyper.NewTree(apiKey, cache, leaves, hashing.NewSha256Hasher())
	return balloon.NewHyperBalloon(hashing.NewSha256Hasher(), history, hyper), nil
}

func newHTTPServer(endpoint string, balloon *balloon.HyperBalloon, signer sign.Signer) *http.Server {
	router := apihttp.NewApiHttp(balloon, signer)
	server := &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(router),
	}
	return server
}

func newProfilingServer(endpoint string) *http.Server {
	server := &http.Server{
		Addr:    endpoint,
		Handler: nil,
	}
	return server
}

func newTamperingServer(endpoint string, store tampering.DeletableStore, hasher hashing.Hasher) *http.Server {
	router := tampering.NewTamperingApi(store, hasher)
	server := &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(router),
	}
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
