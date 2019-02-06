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
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// HyperStats has a Map of all the stats relative to our Hyper Tree
	Hyper *expvar.Map
	// HistoryStats has a Map of all the stats relative to our History Tree
	History *expvar.Map
	// Qed_balloonStats has a Map of all the stats relative to Qed_balloon
	Qed_balloon *expvar.Map

	// Prometheus
	// API
	Qed_api_healthcheck_requests_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "API_healthcheck_requests_total",
			Help: "The total number of healthcheck api requests",
		},
	)

	// Qed_balloon
	Qed_Qed_balloon_add_duration_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_balloon_add_duration_seconds",
			Help: "Duration of the 'Add' Qed_balloon method.",
		},
	)

	Qed_balloon_membership_duration_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_balloon_membership_duration_seconds",
			Help: "Duration of the 'Membership' Qed_balloon method.",
		},
	)

	Qed_balloon_digest_membership_duration_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_balloon_digest_membership_duration_seconds",
			Help: "Duration of the 'Digest Membership' Qed_balloon method",
		},
	)

	Qed_balloon_incremental_duration_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_balloon_incremental_duration_seconds",
			Help: "Duration of the 'Incremental' Qed_balloon method.",
		},
	)

	Qed_balloon_add_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_balloon_add_total",
			Help: "Amount of 'Add' operation API calls.",
		},
	)

	Qed_balloon_membership_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_balloon_membership_total",
			Help: "Amount of 'Membership' operation API calls.",
		},
	)

	Qed_balloon_digest_membership_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_balloon_digest_membership_total",
			Help: "Amount of 'Membership by digest' operation API calls.",
		},
	)

	Qed_balloon_incremental_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_balloon_incremental_total",
			Help: "Amount of 'Incremental' operation API calls.",
		},
	)

	// Agents
	Qed_sender_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_sender_instances_count",
			Help: "Amount of Qed_sender agents instanciated",
		},
	)

	Qed_auditor_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_auditor_instances_count",
			Help: "Amount of Qed_auditor agents instanciated",
		},
	)

	Qed_monitor_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_monitor_instances_count",
			Help: "Amount of Qed_monitor agents instanciated",
		},
	)

	Qed_publisher_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_publisher_instances_count",
			Help: "Amount of Qed_publisher agents instanciated.",
		},
	)

	Qed_sender_batches_sent_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_sender_batches_sent_total",
			Help: "Amount of batches sent by Sender.",
		},
	)

	Qed_auditor_batches_received_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_auditor_batches_received_total",
			Help: "Amount of batches received by Auditor.",
		},
	)

	Qed_monitor_batches_received_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_monitor_batches_received_total",
			Help: "Amount of batches received by Monitor.",
		},
	)

	Qed_publisher_batches_received_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "Qed_publisher_batches_received_total",
			Help: "Amount of batches received by Publisher.",
		},
	)

	Qed_auditor_batches_process_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_auditor_batches_process_seconds",
			Help: "Duration of Auditor batch processing",
		},
	)

	Qed_monitor_batches_process_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_monitor_batches_process_seconds",
			Help: "Duration of Monitor batch processing",
		},
	)

	Qed_publisher_batches_process_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "Qed_publisher_batches_process_seconds",
			Help: "Duration of Publisher batch processing",
		},
	)

	// Example
	Qed_exampleGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "example_function_duration_seconds",
			Help: "Duration of the last call of an example function.",
		},
	)

	Qed_exampleSummary = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "example_function_durations_seconds",
			Help: "example function latency distributions.",
		},
	)

	Qed_exampleHistogram = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "example_request_duration_seconds",
			Help:    "Histogram for the runtime of a simple example function.",
			Buckets: prometheus.LinearBuckets(0.01, 0.01, 10),
		},
	)

	metricsList = []prometheus.Collector{
		Qed_api_healthcheck_requests_total,

		Qed_balloon_add_duration_seconds,
		Qed_balloon_membership_duration_seconds,
		Qed_balloon_digest_membership_duration_seconds,
		Qed_balloon_incremental_duration_seconds,

		Qed_balloon_add_total,
		Qed_balloon_membership_total,
		Qed_balloon_digest_membership_total,
		Qed_balloon_incremental_total,

		Qed_sender_instances_count,
		Qed_auditor_instances_count,
		Qed_monitor_instances_count,
		Qed_publisher_instances_count,

		Qed_sender_batches_sent_total,
		Qed_auditor_batches_received_total,
		Qed_monitor_batches_received_total,
		Qed_publisher_batches_received_total,

		Qed_auditor_batches_process_seconds,
		Qed_monitor_batches_process_seconds,
		Qed_publisher_batches_process_seconds,

		Qed_exampleGauge,
		Qed_exampleSummary,
		Qed_exampleHistogram,
	}
)

var registerMetrics sync.Once

// Register all metrics.
func Register(r *prometheus.Registry) {
	// Register the metrics.
	registerMetrics.Do(
		func() {
			for _, metric := range metricsList {
				r.MustRegister(metric)
			}
		},
	)
}

// Implement expVar.Var interface
type Uint64ToVar uint64

func (v Uint64ToVar) String() string {
	return fmt.Sprintf("%d", v)
}

func init() {
	Hyper = expvar.NewMap("hyper_stats")
	History = expvar.NewMap("history_stats")
	Qed_balloon = expvar.NewMap("Qed_balloon_stats")
}
