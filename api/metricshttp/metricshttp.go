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

// Package metricshttp implements the Metrics HTTP API public interface.
package metricshttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// NewMetricsHTTP will return a mux server with the endpoint required to
// look for metrics, which are retrieved using Prometheus.
func NewMetricsHTTP(r *prometheus.Registry) *http.ServeMux {
	mux := http.NewServeMux()
	g := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		r,
	}

	handler := promhttp.HandlerFor(g, promhttp.HandlerOpts{})
	instrumentedHandler := promhttp.InstrumentMetricHandler(r, handler)
	mux.Handle("/metrics", instrumentedHandler)
	return mux
}
