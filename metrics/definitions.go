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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (

	// API

	QedAPIHealthcheckRequestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_api_healthcheck_requests_total",
			Help: "The total number of healthcheck api requests",
		},
	)

	// BALLOON

	QedBalloonAddDurationSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_add_duration_seconds",
			Help: "Duration of the add operation.",
		},
	)
	QedBalloonMembershipDurationSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_membership_duration_seconds",
			Help: "Duration of the membership queries.",
		},
	)
	QedBalloonDigestMembershipDurationSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_digest_membership_duration_seconds",
			Help: "Duration of the membership by digest queries.",
		},
	)
	QedBalloonIncrementalDurationSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_balloon_incremental_duration_seconds",
			Help: "Duration of the incremental queries.",
		},
	)
	QedBalloonAddTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_balloon_add_total",
			Help: "Number of add operations",
		},
	)
	QedBalloonMembershipTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_balloon_membership_total",
			Help: "Number of membership queries.",
		},
	)
	QedBalloonDigestMembershipTotal = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_balloon_digest_membership_total",
			Help: "Number of membership by digest queries.",
		},
	)
	QedBalloonIncrementalTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_balloon_incremental_total",
			Help: "Number of incremental queries.",
		},
	)

	// HYPER TREE

	QedHyperAddTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_hyper_add_total",
			Help: "Number of the events added to the hyper tree.",
		},
	)
	QedHyperMembershipTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_hyper_membership_total",
			Help: "Number of membership queries",
		},
	)

	// HISTORY TREE

	QedHistoryAddTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_history_add_total",
			Help: "Number of the events added to the history tree.",
		},
	)
	QedHistoryMembershipTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_history_membership_total",
			Help: "Number of membership queries",
		},
	)
	QedHistoryConsistencyTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_history_consistency_total",
			Help: "Number of consistency queries",
		},
	)

	// PROMETHEUS

	DefaultMetrics = []prometheus.Collector{
		QedAPIHealthcheckRequestsTotal,

		QedBalloonAddDurationSeconds,
		QedBalloonMembershipDurationSeconds,
		QedBalloonDigestMembershipDurationSeconds,
		QedBalloonIncrementalDurationSeconds,

		QedBalloonAddTotal,
		QedBalloonMembershipTotal,
		QedBalloonDigestMembershipTotal,
		QedBalloonIncrementalTotal,
	}
)
