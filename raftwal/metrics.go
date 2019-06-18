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

package raftwal

import "github.com/prometheus/client_golang/prometheus"

// namespace is the leading part of all published metrics.
const namespace = "qed"

// subsystem associated with metrics for raft balloon
const subSystem = "raft_balloon"

// raftBalloonMetrics is the definition of the set of metrics retrieved in raft.
type raftBalloonMetrics struct {
	Version                 prometheus.GaugeFunc
	Adds                    prometheus.Counter
	MembershipQueries       prometheus.Counter
	DigestMembershipQueries prometheus.Counter
	IncrementalQueries      prometheus.Counter
}

func newRaftBalloonMetrics(b *RaftBalloon) *raftBalloonMetrics {
	return &raftBalloonMetrics{
		Version: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystem,
				Name:      "version",
				Help:      "Current balloon version.",
			},
			func() float64 {
				return float64(b.fsm.balloon.Version() - 1)
			},
		),
		Adds: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystem,
				Name:      "adds",
				Help:      "Number of add operations",
			},
		),
		MembershipQueries: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystem,
				Name:      "membership_queries",
				Help:      "Number of membership queries.",
			},
		),
		DigestMembershipQueries: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: subSystem,
				Name:      "digest_membership_queries",
				Help:      "Number of membership by digest queries.",
			},
		),
		IncrementalQueries: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: subSystem,
				Name:      "incremental_queries",
				Help:      "Number of incremental queries.",
			},
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *raftBalloonMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.Version,
		m.Adds,
		m.MembershipQueries,
		m.DigestMembershipQueries,
		m.IncrementalQueries,
	}
}
