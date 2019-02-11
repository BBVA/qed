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
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Prometheus
	// Agents
	Qed_auditor_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_auditor_instances_count",
			Help: "Amount of Qed_auditor agents instanciated",
		},
	)

	Qed_monitor_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_monitor_instances_count",
			Help: "Amount of Qed_monitor agents instanciated",
		},
	)

	Qed_publisher_instances_count = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_publisher_instances_count",
			Help: "Amount of Qed_publisher agents instanciated.",
		},
	)

	Qed_auditor_batches_received_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_auditor_batches_received_total",
			Help: "Amount of batches received by Auditor.",
		},
	)

	Qed_monitor_batches_received_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_monitor_batches_received_total",
			Help: "Amount of batches received by Monitor.",
		},
	)

	Qed_publisher_batches_received_total = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_publisher_batches_received_total",
			Help: "Amount of batches received by Publisher.",
		},
	)

	Qed_auditor_batches_process_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_auditor_batches_process_seconds",
			Help: "Duration of Auditor batch processing",
		},
	)

	Qed_monitor_batches_process_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_monitor_batches_process_seconds",
			Help: "Duration of Monitor batch processing",
		},
	)

	Qed_publisher_batches_process_seconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_publisher_batches_process_seconds",
			Help: "Duration of Publisher batch processing",
		},
	)

	metricsList = []prometheus.Collector{
		Qed_auditor_instances_count,
		Qed_monitor_instances_count,
		Qed_publisher_instances_count,

		Qed_auditor_batches_received_total,
		Qed_monitor_batches_received_total,
		Qed_publisher_batches_received_total,

		Qed_auditor_batches_process_seconds,
		Qed_monitor_batches_process_seconds,
		Qed_publisher_batches_process_seconds,
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
