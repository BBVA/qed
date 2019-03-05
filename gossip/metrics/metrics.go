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
	QedAuditorInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_auditor_instances_count",
			Help: "Number of auditor agents running.",
		},
	)

	QedMonitorInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_monitor_instances_count",
			Help: "Number of monitor agents running.",
		},
	)

	QedPublisherInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_publisher_instances_count",
			Help: "Number of publisher agents running.",
		},
	)

	QedAuditorBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_auditor_batches_received_total",
			Help: "Number of batches received by auditors.",
		},
	)

	QedMonitorBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_monitor_batches_received_total",
			Help: "Number of batches received by monitors.",
		},
	)

	QedPublisherBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_publisher_batches_received_total",
			Help: "Number of batches received by publishers.",
		},
	)

	QedAuditorBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_auditor_batches_process_seconds",
			Help: "Duration of Auditor batch processing",
		},
	)

	QedMonitorBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_monitor_batches_process_seconds",
			Help: "Duration of Monitor batch processing",
		},
	)

	QedPublisherBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_publisher_batches_process_seconds",
			Help: "Duration of Publisher batch processing",
		},
	)

	QedAuditorGetMembershipProofErrTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_auditor_get_membership_proof_err_total",
			Help: "Number of errors trying to get membership proofs by auditors.",
		},
	)

	metricsList = []prometheus.Collector{
		QedAuditorInstancesCount,
		QedMonitorInstancesCount,
		QedPublisherInstancesCount,

		QedAuditorBatchesReceivedTotal,
		QedMonitorBatchesReceivedTotal,
		QedPublisherBatchesReceivedTotal,

		QedAuditorBatchesProcessSeconds,
		QedMonitorBatchesProcessSeconds,
		QedPublisherBatchesProcessSeconds,

		QedAuditorGetMembershipProofErrTotal,
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
