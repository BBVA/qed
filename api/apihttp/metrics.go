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

package apihttp

import (
	"github.com/bbva/qed/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

// namespace is the leading part of all published metrics.
const namespace = "qed"

// subsystem associated with metrics for API HTTP
const subSystem = "api_http"

type raftNodeMetrics struct {
	Hits prometheus.GaugeFunc
}

var (
	HealthCheckRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "health_check_requests",
			Help:      "Number of current HTTP HealtCheck requests.",
		},
	)
	AddRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "add_requests",
			Help:      "Number of current HTTP Add requests.",
		},
	)
	AddBulkRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "add_bulk_requests",
			Help:      "Number of current HTTP AddBulk requests.",
		},
	)
	MembershipRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "membership_requests",
			Help:      "Number of current HTTP Membership requests.",
		},
	)
	DigestMembershipRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "digest_menbership_requests",
			Help:      "Number of HTTP Digest Membreship requests.",
		},
	)
	IncrementalRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "incremental_requests",
			Help:      "Number of current HTTP Incremental requests.",
		},
	)
	InfoRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "info_requests",
			Help:      "Number of current HTTP Info requests.",
		},
	)
	InfoShardsRequest = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "info_shards_requests",
			Help:      "Number of current HTTP Info Shards requests.",
		},
	)
)

func RegisterMetrics(registry metrics.Registry) {
	if registry != nil {
		registry.MustRegister(
			HealthCheckRequest,
			AddRequest,
			AddBulkRequest,
			MembershipRequest,
			DigestMembershipRequest,
			IncrementalRequest,
			InfoRequest,
			InfoShardsRequest,
		)
	}
}
