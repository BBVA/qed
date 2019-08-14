/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/api/mgmthttp"
	"github.com/bbva/qed/consensus"
	"github.com/bbva/qed/crypto/sign"
	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log2"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/storage/rocks"
	"github.com/prometheus/client_golang/prometheus"
)

// Server encapsulates the data and login to start/stop a QED server
type Server struct {
	conf               *Config
	bootstrap          bool // Set bootstrap to true when bringing up the first node as a master
	httpServer         *http.Server
	mgmtServer         *http.Server
	raftNode           *consensus.RaftNode
	metrics            *serverMetrics
	metricsServer      *metrics.Server
	prometheusRegistry *prometheus.Registry
	signer             sign.Signer
	sender             *Sender
	agent              *gossip.Agent
	snapshotsCh        chan *protocol.Snapshot
	log                log2.Logger
}

// NewServer creates a new Server based on the parameters it receives.
func NewServer(conf *Config) (*Server, error) {
	return NewServerWithLogger(conf, log2.Default())
}

// NewServerWithLogger creates a new Server based on the parameters it receives and
// configures a logger.
func NewServerWithLogger(conf *Config, logger log2.Logger) (*Server, error) {

	bootstrap := false
	if len(conf.RaftJoinAddr) <= 0 {
		bootstrap = true
	}

	server := &Server{
		conf:      conf,
		bootstrap: bootstrap,
		log:       logger,
	}

	logger.Infof("Ensuring directory at %s exists", conf.DBPath)
	if err := os.MkdirAll(conf.DBPath, 0755); err != nil {
		return nil, err
	}

	logger.Infof("Ensuring directory at %s exists", conf.RaftPath)
	if err := os.MkdirAll(conf.RaftPath, 0755); err != nil {
		return nil, err
	}

	// Open RocksDB store
	store, err := rocks.NewRocksDBStore(conf.DBPath, conf.DbWalTtl)
	if err != nil {
		return nil, err
	}

	// Create signer
	server.signer, err = sign.NewEd25519SignerFromFile(conf.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	// Create metrics server
	server.metricsServer = metrics.NewServer(conf.MetricsAddr)

	// Create profiling server
	if server.conf.EnableProfiling {
		go func() {
			logger.Infof("\t* Starting QED Profiling server in addr: %s", server.conf.ProfilingAddr)
			err := http.ListenAndServe(server.conf.ProfilingAddr, nil)
			if err != http.ErrServerClosed {
				logger.Fatalf("Can't start QED Profiling Server: %v", err)
			}
		}()
	}

	// Create gossip agent
	config := gossip.DefaultConfig()
	config.BindAddr = conf.GossipAddr
	config.Role = "server"
	config.NodeName = conf.NodeID

	server.agent, err = gossip.NewAgentFromConfigWithLogger(config, server.log.Named("sender.agent"))
	if err != nil {
		return nil, err
	}

	// TODO: add queue size to config
	server.snapshotsCh = make(chan *protocol.Snapshot, 1<<16)

	// Create sender
	server.sender = NewSenderWithLogger(server.agent, server.signer, 500, 2, 3, server.log.Named("sender"))

	// Create RaftBalloon
	clusterOpts := consensus.DefaultClusteringOptions()
	clusterOpts.NodeID = conf.NodeID
	clusterOpts.Addr = conf.RaftAddr
	clusterOpts.HttpAddr = conf.HTTPAddr
	clusterOpts.RaftLogPath = conf.RaftPath
	clusterOpts.MgmtAddr = conf.MgmtAddr
	clusterOpts.Bootstrap = bootstrap
	clusterOpts.RaftLogging = true
	if !bootstrap {
		clusterOpts.Seeds = conf.RaftJoinAddr
	}
	server.raftNode, err = consensus.NewRaftNodeWithLogger(clusterOpts, store, server.snapshotsCh, server.log.Named("cluster"))
	if err != nil {
		return nil, err
	}

	// Create http endpoints
	httpMux := apihttp.NewApiHttp(server.raftNode)
	if conf.EnableTLS {
		server.httpServer = newTLSServer(conf.HTTPAddr, httpMux, logger.Named("api"))
	} else {
		server.httpServer = newHTTPServer(conf.HTTPAddr, httpMux, logger.Named("api"))
	}

	// Create management endpoints
	mgmtMux := mgmthttp.NewMgmtHttp(server.raftNode)
	server.mgmtServer = newHTTPServer(conf.MgmtAddr, mgmtMux, logger.Named("mgmt"))

	// register qed metrics
	server.metrics = newServerMetrics()
	apihttp.RegisterMetrics(server.metricsServer)
	server.RegisterMetrics(server.metricsServer)
	store.RegisterMetrics(server.metricsServer)
	server.raftNode.RegisterMetrics(server.metricsServer)
	server.sender.RegisterMetrics(server.metricsServer)

	return server, nil
}

// Start will start the server in a non-blockable fashion.
func (s *Server) Start() error {
	s.metrics.Instances.Inc()
	s.log.Infof("Starting QED server. Node ID: %s", s.conf.NodeID)

	metadata := map[string]string{}
	metadata["HTTPAddr"] = s.conf.HTTPAddr

	s.log.Infof("\t* Starting metrics HTTP server in addr: %s", s.conf.MetricsAddr)
	go func() {
		if err := s.metricsServer.Start(); err != http.ErrServerClosed {
			s.log.Fatalf("Can't start metrics HTTP server: %s", err)
		}
	}()

	// TODO remove goroutines

	if s.conf.EnableTLS {
		go func() {
			s.log.Infof("\t* Starting QED API HTTPS server in addr: %s", s.conf.HTTPAddr)
			err := s.httpServer.ListenAndServeTLS(
				s.conf.SSLCertificate,
				s.conf.SSLCertificateKey,
			)
			if err != http.ErrServerClosed {
				s.log.Fatalf("Can't start QED API HTTP Server: %v", err) // TODO should we return error instead of exiting?
			}
		}()
	} else {
		go func() {
			s.log.Infof("\t* Starting QED API HTTP server in addr: %s", s.conf.HTTPAddr)
			if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
				s.log.Fatalf("Can't start QED API HTTP Server: %v", err)
			}
		}()
	}

	go func() {
		s.log.Infof("\t* Starting QED MGMT HTTP server in addr: %s", s.conf.MgmtAddr)
		if err := s.mgmtServer.ListenAndServe(); err != http.ErrServerClosed {
			s.log.Fatalf("Can't start QED MGMT HTTP Server: %v", err)
		}
	}()

	s.log.Info("Starting QED agent...")
	s.agent.Start()

	s.log.Info("Starting snapshots sender...")
	s.sender.Start(s.snapshotsCh)

	if err := s.raftNode.WaitForLeader(5 * time.Second); err != nil {
		return err
	}

	s.log.Infof("Server ready on %s and %s", s.conf.HTTPAddr, s.conf.MgmtAddr)

	return nil
}

// Stop will close all the channels from the mux servers.
func (s *Server) Stop() error {
	s.metrics.Instances.Dec()
	s.log.Infof("Shutting down QED server. Node ID: %s", s.conf.NodeID)

	s.log.Info("Metrics enabled: stopping server...")
	s.metricsServer.Shutdown()

	s.log.Info("Stopping MGMT server...")
	if err := s.mgmtServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		s.log.Errorf("Unable to stop MGMT server: %v", err)
		return err
	}

	s.log.Info("Stopping API HTTP server...")
	if err := s.httpServer.Shutdown(context.Background()); err != nil { // TODO include timeout instead nil
		s.log.Errorf("Unable to stop API HTTP server: %v", err)
		return err
	}

	s.log.Info("Closing QED sender...")
	s.sender.Stop()

	s.log.Info("Stopping QED agent...")
	if err := s.agent.Shutdown(); err != nil {
		s.log.Errorf("Unable to stop agent %v", err)
		return err
	}

	s.log.Info("Stopping RAFT node...")
	err := s.raftNode.Close(true)
	if err != nil {
		s.log.Errorf("Unable to stop raft node: %v", err)
		return err
	}

	close(s.snapshotsCh)

	s.log.Info("Done. Exiting...")
	return nil
}

func (s *Server) RegisterMetrics(registry metrics.Registry) {
	if registry != nil {
		registry.MustRegister(s.metrics.collectors()...)
	}
}

func newTLSServer(addr string, mux *http.ServeMux, logger log2.Logger) *http.Server {

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{
			tls.CurveP521,
			tls.CurveP384,
			tls.CurveP256,
		},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	return &http.Server{
		Addr:         addr,
		Handler:      apihttp.LogHandler(mux, logger),
		TLSConfig:    cfg,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0),
	}

}

func newHTTPServer(addr string, mux *http.ServeMux, logger log2.Logger) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: apihttp.LogHandler(mux, logger),
	}
}
