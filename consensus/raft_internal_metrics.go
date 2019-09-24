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

package consensus

import (
	"strconv"

	"github.com/hashicorp/raft"
	"github.com/prometheus/client_golang/prometheus"
)

// namespace is the leading part of all Raft published metrics.
const raftNamespace = "qed"

// subsystem associated with Raft internal metrics.
const raftSubSystem = "raft_internal"

type raftInternalMetrics struct {
	LastSnapshotIndex prometheus.GaugeFunc
	LastSnapshotTerm  prometheus.GaugeFunc
	CommitIndex       prometheus.GaugeFunc
	AppliedIndex      prometheus.GaugeFunc
	FsmPending        prometheus.GaugeFunc
	LastLogIndex      prometheus.GaugeFunc
	LastLogTerm       prometheus.GaugeFunc
}

func newRaftInternalMetrics(raft *raft.Raft) *raftInternalMetrics {
	return &raftInternalMetrics{
		LastLogIndex: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: raftNamespace,
				Subsystem: raftSubSystem,
				Name:      "last_log_index",
				Help:      "Last log index.",
			},
			func() float64 {
				stats := raft.Stats()
				index, _ := strconv.ParseFloat(stats["last_log_index"], 64)
				return index
			},
		),
		LastLogTerm: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: raftNamespace,
				Subsystem: raftSubSystem,
				Name:      "last_log_term",
				Help:      "Last log term.",
			},
			func() float64 {
				stats := raft.Stats()
				term, _ := strconv.ParseFloat(stats["last_log_term"], 64)
				return term
			},
		),
		CommitIndex: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: raftNamespace,
				Subsystem: raftSubSystem,
				Name:      "commit_index",
				Help:      "Commit index.",
			},
			func() float64 {
				stats := raft.Stats()
				index, _ := strconv.ParseFloat(stats["commit_index"], 64)
				return index
			},
		),
		AppliedIndex: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: raftNamespace,
				Subsystem: raftSubSystem,
				Name:      "applied_index",
				Help:      "Applied index.",
			},
			func() float64 {
				stats := raft.Stats()
				index, _ := strconv.ParseFloat(stats["applied_index"], 64)
				return index
			},
		),
		FsmPending: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: raftNamespace,
				Subsystem: raftSubSystem,
				Name:      "fsm_pending",
				Help:      "Fsm pending.",
			},
			func() float64 {
				stats := raft.Stats()
				pending, _ := strconv.ParseFloat(stats["fsm_pending"], 64)
				return pending
			},
		),
		LastSnapshotIndex: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: raftNamespace,
				Subsystem: raftSubSystem,
				Name:      "last_snapshot_index",
				Help:      "Last snapshot index.",
			},
			func() float64 {
				stats := raft.Stats()
				index, _ := strconv.ParseFloat(stats["last_snapshot_index"], 64)
				return index
			},
		),
		LastSnapshotTerm: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: raftNamespace,
				Subsystem: raftSubSystem,
				Name:      "last_snapshot_term",
				Help:      "Last snapshot term.",
			},
			func() float64 {
				stats := raft.Stats()
				term, _ := strconv.ParseFloat(stats["last_snapshot_term"], 64)
				return term
			},
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *raftInternalMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.LastLogIndex,
		m.LastLogTerm,
		m.CommitIndex,
		m.AppliedIndex,
		m.FsmPending,
		m.LastSnapshotIndex,
		m.LastSnapshotTerm,
	}
}
