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

package metrics

import (
	"expvar"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HyperStats has a Map of all the stats relative to our Hyper Tree
	Hyper *expvar.Map
	// HistoryStats has a Map of all the stats relative to our History Tree
	History *expvar.Map
	// BalloonStats has a Map of all the stats relative to Balloon
	Balloon *expvar.Map

	// Prometheus
	FuncDuration    prometheus.Gauge
	RequestDuration prometheus.Histogram
	OpsProcessed    prometheus.Counter
)

// Implement expVar.Var interface
type Uint64ToVar uint64

func (v Uint64ToVar) String() string {
	return fmt.Sprintf("%d", v)
}

func init() {
	Hyper = expvar.NewMap("hyper_stats")
	History = expvar.NewMap("history_stats")
	Balloon = expvar.NewMap("balloon_stats")

	FuncDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "example_function_duration_seconds",
		Help: "Duration of the last call of an example function.",
	})

	RequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "example_request_duration_seconds",
		Help:    "Histogram for the runtime of a simple example function.",
		Buckets: prometheus.LinearBuckets(0.01, 0.01, 10),
	})

	OpsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "qed_healthcheck_ops_total",
		Help: "The total number of processed events",
	})

}
