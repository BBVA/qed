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

// Package metrics implements an HTTP metrics server along with its
// life cycle: start, shutdown, and register metrics.
package metrics

import (
	"context"
	"net/http"
	"time"

	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Registry defines the interface of the prometheus metrics registry
type Registry interface {
	Register(prometheus.Collector) error
	MustRegister(...prometheus.Collector)
}

type Registerer interface {
	RegisterMetrics(Registry)
}

// A metrics server holds the http API and the prometheus registry
// which provides access to the registered metrics.
type Server struct {
	server   *http.Server
	registry *prometheus.Registry

	log log.Logger
}

// Create new metrics server. Do not listen to the given address until
// the server is started.
func NewServer(addr string) *Server {
	r := prometheus.NewRegistry()
	return &Server{
		server: &http.Server{
			Addr:    addr,
			Handler: metricshttp.NewMetricsHTTP(r),
		},
		registry: r,
	}
}

// Listens on the configured address and blocks until shutdown is called.
func (m Server) Start() error {
	return m.server.ListenAndServe()
}

// Shutdown gracefully shutdowns metrics http server waiting 5 seconds for
// connections to be closed.
func (m Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	m.server.Shutdown(ctx)
}

// Register registers a prometheus collector in the prometheus registry used
// by the metrics server.
func (m Server) Register(collector prometheus.Collector) error {
	return m.registry.Register(collector)
}

// MustRegister registers multiple prometheus collectors in the prometheus registry used
// by the metrics server.
func (m Server) MustRegister(collectors ...prometheus.Collector) {
	m.registry.MustRegister(collectors...)
}
