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
	"net/http"
	_ "net/http/pprof" // this will enable the default profiling capabilities
	"os"
	"os/signal"
	"syscall"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/api/tampering"
	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/sign"
	"github.com/bbva/qed/storage"
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
	nodeID         string // unique name for node. If not set, fallback to hostname
	httpAddr       string // HTTP server bind address
	raftAddr       string // Raft communication bind address
	mgmtAddr       string // Management server bind address
	joinAddr       string // Comma-delimited list of nodes, through wich a cluster can be joined (protocol://host:port)
	dbPath         string // Path to storage directory
	raftPath       string // Path to Raft storage directory
	privateKeyPath string // Path to the private key file used to sign commitments
	apiKey         string

	httpServer      *http.Server
	mgmtServer      *http.Server
	raftBalloon     *balloon.RaftBalloon
	tamperingServer *http.Server
	profilingServer *http.Server
	signer          sign.Signer
}

// NewServer synthesizes a new Server based on the parameters it receives.
func NewServer(
	nodeID string,
	httpAddr string,
	raftAddr string,
	mgmtAddr string,
	joinAddr string,
	dbPath string,
	raftPath string,
	privateKeyPath string,
	apiKey string,
	enableProfiling bool,
	enableTampering bool,
) (*Server, error) {

	server := &Server{
		nodeID:   nodeID,
		httpAddr: httpAddr,
		raftAddr: raftAddr,
		mgmtAddr: mgmtAddr,
		joinAddr: joinAddr,
		dbPath:   dbPath,
		raftPath: raftPath,
		apiKey:   apiKey,
	}

	log.Infof("ensuring directory at %s exists", dbPath)
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, err
	}

	log.Infof("ensuring directory at %s exists", raftPath)
	if err := os.MkdirAll(raftPath, 0755); err != nil {
		return nil, err
	}

	// Open badger store
	store, err := badger.NewBadgerStore(dbPath)
	if err != nil {
		return nil, err
	}

	// Create signer
	server.signer, err = sign.NewEd25519SignerFromFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	// Create RaftBalloon
	server.raftBalloon, err = balloon.NewRaftBalloon(raftPath, raftAddr, nodeID, store)
	if err != nil {
		return nil, err
	}

	// Create http endpoints
	server.httpServer = newHTTPServer(server.httpAddr, server.raftBalloon, server.signer)
	//server.mgmtServer = newMgmtServer(server.mgmtAddr, server.raftBalloon)
	if enableTampering {
		server.tamperingServer = newTamperingServer("localhost:8081", store, hashing.NewSha256Hasher())
	}
	if enableProfiling {
		server.profilingServer = newProfilingServer("localhost:6060")
	}

	return server, nil

}

// Start will start the server in a non-blockable fashion.
func (s *Server) Start() error {

	log.Debugf("Starting QED server...")

	err := s.raftBalloon.Open(true)
	if err != nil {
		return err
	}

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
		log.Debug("	* Starting QED API HTTP server in addr: ", s.httpAddr)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start QED API HTTP Server: %s", err)
		}
	}()

	log.Debugf(" ready on %s\n", s.httpAddr)

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

	log.Debugf("Stopping RAFT server...")
	err := s.raftBalloon.Close(true)
	if err != nil {
		log.Error(err)
	}
	log.Debugf("Done. Exiting...\n")
}

func newHTTPServer(endpoint string, raftBalloon balloon.RaftBalloonApi, signer sign.Signer) *http.Server {
	router := apihttp.NewApiHttp(raftBalloon, signer)
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

func newTamperingServer(endpoint string, store storage.DeletableStore, hasher hashing.Hasher) *http.Server {
	router := tampering.NewTamperingApi(store, hasher)
	server := &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(router),
	}
	return server
}

func awaitTermSignal(closeFn func()) {

	signals := make(chan os.Signal, 1)
	// sigint: Ctrl-C, sigterm: kill command
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// block main and wait for a signal
	sig := <-signals
	log.Infof("Signal received: %v", sig)

	closeFn()
}
