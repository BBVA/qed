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

package gossip

import (
	"context"
	"net/http"
	"time"

	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/log"
	"github.com/prometheus/client_golang/prometheus"
)

type metricsServer struct {
	server   *http.Server
	registry *prometheus.Registry
}

func newMetricsServer(addr string) *metricsServer {
	r := prometheus.NewRegistry()
	return &metricsServer{
		server: &http.Server{
			Addr:    addr,
			Handler: metricshttp.NewMetricsHTTP(r),
		},
		registry: r,
	}
}

func (m metricsServer) start() {
	if err := m.server.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Can't start metrics HTTP server: %s", err)
	}
}

func (m metricsServer) shutdown() {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	m.server.Shutdown(ctx)
}

func (m metricsServer) register(metric prometheus.Collector) {
	if err := m.registry.Register(metric); err != nil {
		log.Infof("metric not registered:", err)
	} else {
		log.Infof("metric registered.")
	}
}

