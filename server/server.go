package server

import (
	"context"
	"net/http"
	"time"
	apihttp "verifiabledata/api/http"
	"verifiabledata/util"

	"github.com/golang/glog"
)

// Server encapsulates the data and logic to start a VD server
type Server struct {
	// Endpoint for REST server
	HTTPEndpoint string

	HTTPServer *http.Server
}

func (s *Server) Run(ctx context.Context) error {

	s.HTTPServer = startHttpServer(s.HTTPEndpoint)

	go util.AwaitTermSignal(s.Stop)

	glog.Infof("Stopping server, about to exit...")

	// Give things a few seconds to tidy up
	time.Sleep(time.Second * 2)

	return nil

}

func startHttpServer(endpoint string) *http.Server {

	glog.Infof("HTTP server starting on %v", endpoint)

	srv := &http.Server{Addr: endpoint}
	http.HandleFunc("/health-check", apihttp.HealthCheckHandler)
	http.Handle("/events", &apihttp.EventInsertHandler{InsertRequestQueue: make(chan *apihttp.InsertRequest)})

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
