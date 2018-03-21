// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file
package server

import (
	"context"
	"github.com/golang/glog"
	"net/http"
	"time"

	apihttp "verifiabledata/api/http"
	"verifiabledata/merkle/history"
	"verifiabledata/sequencer"
	"verifiabledata/util"
)

// Server encapsulates the data and logic to start a VD server
type Server struct {
	// Endpoint for REST server
	HTTPEndpoint string

	HTTPServer *http.Server
}

func (s *Server) Run(ctx context.Context) error {

	s.HTTPServer = startHttpServer(s.HTTPEndpoint)

	util.AwaitTermSignal(s.Stop)

	glog.Infof("Stopping server, about to exit...")

	// Give things a few seconds to tidy up
	time.Sleep(time.Second * 2)

	return nil

}

func startHttpServer(endpoint string) *http.Server {

	glog.Infof("HTTP server starting on %v", endpoint)

	fetchQueue := make(chan *apihttp.FetchRequest)

	// INFO: Creating HistoryTree for now. We will need a process to subscribe
	// to a external one in the distributed future
	tree := history.NewInmemoryTree()
	seq := sequencer.NewSequencer(1000, tree)

	seq.Start()

	srv := &http.Server{Addr: endpoint}
	http.Handle("/health-check", apihttp.AuthHandlerMiddleware(apihttp.HealthCheckHandler))
	http.Handle("/events", apihttp.AuthHandlerMiddleware(apihttp.QueueHandlerConstructor(seq.InsertRequestQueue)))
	http.Handle("/fetch", apihttp.AuthHandlerMiddleware(apihttp.FetchEvent(fetchQueue)))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			glog.Fatalf("HTTPserver: ListenAndServe() error: %s", err)
		}
	}()

	return srv
}

func (s *Server) Stop() {
	glog.Infof("main: stopping HTTP server")
	if err := s.HTTPServer.Shutdown(nil); err != nil { // TODO include timeout instead nil
		panic(err)
	}
	glog.Infof("main: done. Exiting...")
}
