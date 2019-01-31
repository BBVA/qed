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
	// API
	Api_healthcheck_requests_total prometheus.Counter

	// Balloon
	Balloon_add_duration_seconds               prometheus.Gauge
	Balloon_membership_duration_seconds        prometheus.Gauge
	Balloon_digest_membership_duration_seconds prometheus.Gauge
	Balloon_incremental_duration_seconds       prometheus.Gauge

	Balloon_add_total               prometheus.Gauge
	Balloon_membership_total        prometheus.Gauge
	Balloon_digest_membership_total prometheus.Gauge
	Balloon_incremental_total       prometheus.Gauge

	// Agents
	Sender_instances_count    prometheus.Gauge
	Auditor_instances_count   prometheus.Gauge
	Monitor_instances_count   prometheus.Gauge
	Publisher_instances_count prometheus.Gauge

	Sender_batches_sent_total        prometheus.Gauge
	Auditor_batches_received_total   prometheus.Gauge
	Monitor_batches_received_total   prometheus.Gauge
	Publisher_batches_received_total prometheus.Gauge

	Auditor_batches_process_seconds   prometheus.Gauge
	Monitor_batches_process_seconds   prometheus.Gauge
	Publisher_batches_process_seconds prometheus.Gauge

	// Example
	FuncDuration    prometheus.Gauge
	RequestDuration prometheus.Histogram
)

// Implement expVar.Var interface
type Uint64ToVar uint64

func (v Uint64ToVar) String() string {
	return fmt.Sprintf("%d", v)
}

func init() {
	Hyper = expvar.NewMap("hyper_stats")
	History = expvar.NewMap("history_stats")
	Balloon = expvar.NewMap("Balloon_stats")

	apiMetrics()
	balloonMetrics()
	agentMetrics()

	FuncDuration = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "example_function_duration_seconds",
		Help: "Duration of the last call of an example function.",
	})

	RequestDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "example_request_duration_seconds",
		Help:    "Histogram for the runtime of a simple example function.",
		Buckets: prometheus.LinearBuckets(0.01, 0.01, 10),
	})
}

func apiMetrics() {
	Api_healthcheck_requests_total = promauto.NewCounter(prometheus.CounterOpts{
		Name: "Api_healthcheck_requests_total",
		Help: "The total number of healthcheck api requests",
	})
}

func balloonMetrics() {

	Balloon_add_duration_seconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_add_duration_seconds",
		Help: "Duration of the 'Add' balloon method.",
	})

	Balloon_membership_duration_seconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_membership_duration_seconds",
		Help: "Duration of the 'Membership' balloon method.",
	})

	Balloon_digest_membership_duration_seconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_digest_membership_duration_seconds",
		Help: "Duration of the 'Digest Membership' balloon method",
	})

	Balloon_incremental_duration_seconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_incremental_duration_seconds",
		Help: "Duration of the 'Incremental' balloon method.",
	})

	Balloon_add_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_add_total",
		Help: "Duration of the last call of an example function.",
	})

	Balloon_membership_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_membership_total",
		Help: "Duration of the last call of an example function.",
	})

	Balloon_digest_membership_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_digest_membership_total",
		Help: "Duration of the last call of an example function.",
	})

	Balloon_incremental_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Balloon_incremental_total",
		Help: "Duration of the last call of an example function.",
	})
}

func agentMetrics() {

	Sender_instances_count = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Sender_instances_count",
		Help: "Duration of the last call of an example function.",
	})

	Auditor_instances_count = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Auditor_instances_count",
		Help: "Duration of the last call of an example function.",
	})

	Monitor_instances_count = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Monitor_instances_count",
		Help: "Duration of the last call of an example function.",
	})

	Publisher_instances_count = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Publisher_instances_count",
		Help: "Duration of the last call of an example function.",
	})

	Sender_batches_sent_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Sender_batches_sent_total",
		Help: "Duration of the last call of an example function.",
	})

	Auditor_batches_received_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Auditor_batches_received_total",
		Help: "Duration of the last call of an example function.",
	})

	Monitor_batches_received_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Monitor_batches_received_total",
		Help: "Duration of the last call of an example function.",
	})

	Publisher_batches_received_total = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Publisher_batches_received_total",
		Help: "Duration of the last call of an example function.",
	})

	Auditor_batches_process_seconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Auditor_batches_process_seconds",
		Help: "Duration of the last call of an example function.",
	})

	Monitor_batches_process_seconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Monitor_batches_process_seconds",
		Help: "Duration of the last call of an example function.",
	})

	Publisher_batches_process_seconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "Publisher_batches_process_seconds",
		Help: "Duration of the last call of an example function.",
	})
}
