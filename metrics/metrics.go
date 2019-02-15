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
	// BalloonStats has a Map of all the stats relative to Balloon
	Balloon *expvar.Map

	// Prometheus

	// QED
	Qed_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_instances_count",
			Help: "Amount of Qeds instanciated",
		},
	)

	// API
	Qed_api_healthcheck_requests_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_api_healthcheck_requests_total",
			Help: "The total number of healthcheck api requests",
		},
	)

	Qed_balloon_add_duration_seconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_add_duration_seconds",
			Help: "Duration of the 'Add' Qed_balloon method.",
		},
	)

	Qed_balloon_membership_duration_seconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_membership_duration_seconds",
			Help: "Duration of the 'Membership' Qed_balloon method.",
		},
	)

	Qed_balloon_digest_membership_duration_seconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_digest_membership_duration_seconds",
			Help: "Duration of the 'Digest Membership' Qed_balloon method.",
		},
	)

	Qed_balloon_incremental_duration_seconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_incremental_duration_seconds",
			Help: "Duration of the 'Incremental' Qed_balloon method.",
		},
	)

	Qed_balloon_add_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_balloon_add_total",
			Help: "Amount of 'Add' operation API calls.",
		},
	)

	Qed_balloon_membership_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_balloon_membership_total",
			Help: "Amount of 'Membership' operation API calls.",
		},
	)

	Qed_balloon_digest_membership_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_balloon_digest_membership_total",
			Help: "Amount of 'Membership by digest' operation API calls.",
		},
	)

	Qed_balloon_incremental_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_balloon_incremental_total",
			Help: "Amount of 'Incremental' operation API calls.",
		},
	)

	// Sender
	Qed_sender_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_sender_instances_count",
			Help: "Amount of Qed_sender agents instanciated",
		},
	)

	Qed_sender_batches_sent_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_sender_batches_sent_total",
			Help: "Amount of batches sent by Sender.",
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
		Qed_instances_count,
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
		Qed_sender_batches_sent_total,

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
	Balloon = expvar.NewMap("Qed_balloon_stats")
}
