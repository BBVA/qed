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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof" // this will enable the default profiling capabilities
	"os"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/api/mgmthttp"
	"github.com/bbva/qed/api/tampering"
	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/gossip/sender"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/raftwal"
	"github.com/bbva/qed/sign"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/util"
)

// Server encapsulates the data and login to start/stop a QED server
type Server struct {
	conf      *Config
	bootstrap bool // Set bootstrap to true when bringing up the first node as a master

	httpServer      *http.Server
	mgmtServer      *http.Server
	raftBalloon     *raftwal.RaftBalloon
	tamperingServer *http.Server
	profilingServer *http.Server
	signer          sign.Signer
	sender          *sender.Sender
	agent           *gossip.Agent
	agentsQueue     chan *protocol.Snapshot
}

// NewServer creates a new Server based on the parameters it receives.
func NewServer(conf *Config) (*Server, error) {

	bootstrap := false
	if len(conf.RaftJoinAddr) <= 0 {
		bootstrap = true
	}

	server := &Server{
		conf:      conf,
		bootstrap: bootstrap,
	}

	log.Infof("ensuring directory at %s exists", conf.DBPath)
	if err := os.MkdirAll(conf.DBPath, 0755); err != nil {
		return nil, err
	}

	log.Infof("ensuring directory at %s exists", conf.RaftPath)
	if err := os.MkdirAll(conf.RaftPath, 0755); err != nil {
		return nil, err
	}

	// Open badger store
	store, err := badger.NewBadgerStoreOpts(&badger.Options{Path: conf.DBPath, ValueLogGC: true})
	if err != nil {
		return nil, err
	}

	// Create signer
	server.signer, err = sign.NewEd25519SignerFromFile(conf.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	// Create gossip agent
	config := gossip.DefaultConfig()
	config.BindAddr = conf.GossipAddr
	config.Role = member.Server
	server.agent, err = gossip.NewAgent(config, nil)
	if err != nil {
		return nil, err
	}

	if len(conf.GossipJoinAddr) > 0 {
		server.agent.Join(conf.GossipJoinAddr)
	}

	// TODO: add queue size to config
	server.agentsQueue = make(chan *protocol.Snapshot, 10000)

	// Create sender
	server.sender = sender.NewSender(server.agent, sender.DefaultConfig(), server.signer)

	// Create RaftBalloon
	server.raftBalloon, err = raftwal.NewRaftBalloon(conf.RaftPath, conf.RaftAddr, conf.NodeID, store, server.agentsQueue)
	if err != nil {
		return nil, err
	}

	// Create http endpoints
	server.httpServer = newHTTPServer(conf.HttpAddr, server.raftBalloon)

	// Create management endpoints
	server.mgmtServer = newMgmtServer(conf.MgmtAddr, server.raftBalloon)

	if conf.EnableTampering {
		server.tamperingServer = newTamperingServer("localhost:8081", store, hashing.NewSha256Hasher())
	}
	if conf.EnableProfiling {
		server.profilingServer = newProfilingServer("localhost:6060")
	}

	return server, nil
}

func join(joinAddr, raftAddr, nodeID string) error {
	b, err := json.Marshal(map[string]string{"addr": raftAddr, "id": nodeID})
	if err != nil {
		return err
	}
	resp, err := http.Post(fmt.Sprintf("http://%s/join", joinAddr), "application-type/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Start will start the server in a non-blockable fashion.
func (s *Server) Start() error {

	log.Debugf("Starting QED server...")

	err := s.raftBalloon.Open(s.bootstrap)
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
		log.Debug("	* Starting QED API HTTP server in addr: ", s.conf.HttpAddr)
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start QED API HTTP Server: %s", err)
		}
	}()

	go func() {
		log.Debug("	* Starting QED MGMT HTTP server in addr: ", s.conf.MgmtAddr)
		if err := s.mgmtServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start QED MGMT HTTP Server: %s", err)
		}
	}()

	log.Debugf(" ready on %s and %s\n", s.conf.HttpAddr, s.conf.MgmtAddr)

	if !s.bootstrap {
		for _, addr := range s.conf.RaftJoinAddr {
			log.Debug("	* Joining existent cluster QED MGMT HTTP server in addr: ", s.conf.MgmtAddr)
			if err := join(addr, s.conf.RaftAddr, s.conf.NodeID); err != nil {
				log.Fatalf("failed to join node at %s: %s", addr, err.Error())
			}
		}
	}

	go func() {
		log.Debug("	* Starting QED agent.")
		s.sender.Start(s.agentsQueue)
	}()

	util.AwaitTermSignal(s.Stop)

	log.Debug("Stopping server, about to exit...")

	return nil
}

// Stop will close all the channels from the mux servers.
func (s *Server) Stop() error {

	if s.tamperingServer != nil {
		log.Debugf("Tampering enabled: stopping server...")
		if err := s.tamperingServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
			log.Error(err)
			return err
		}
		log.Debugf("Done.\n")
	}

	if s.profilingServer != nil {
		log.Debugf("Profiling enabled: stopping server...")
		if err := s.profilingServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
			log.Error(err)
			return err
		}
		log.Debugf("Done.\n")
	}

	log.Debugf("Stopping MGMT server...")
	if err := s.mgmtServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		log.Error(err)
		return err
	}

	log.Debugf("Stopping HTTP server...")
	if err := s.httpServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		log.Error(err)
		return err
	}

	log.Debugf("Stopping RAFT server...")
	err := s.raftBalloon.Close(true)
	if err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Closing QED sender...")
	s.sender.Stop()
	close(s.agentsQueue)

	log.Debugf("Stopping QED agent...")
	if err := s.agent.Shutdown(); err != nil {
		log.Error(err)
		return err
	}

	log.Debugf("Done. Exiting...\n")
	return nil
}

func newHTTPServer(endpoint string, raftBalloon raftwal.RaftBalloonApi) *http.Server {
	router := apihttp.NewApiHttp(raftBalloon)
	return &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(router),
	}
}

func newMgmtServer(endpoint string, raftBalloon raftwal.RaftBalloonApi) *http.Server {
	router := mgmthttp.NewMgmtHttp(raftBalloon)
	return &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(router),
	}
}

func newProfilingServer(endpoint string) *http.Server {
	return &http.Server{
		Addr:    endpoint,
		Handler: nil,
	}
}

func newTamperingServer(endpoint string, store storage.DeletableStore, hasher hashing.Hasher) *http.Server {
	router := tampering.NewTamperingApi(store, hasher)
	return &http.Server{
		Addr:    endpoint,
		Handler: apihttp.LogHandler(router),
	}
}
