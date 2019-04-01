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

package server

import "github.com/prometheus/client_golang/prometheus"

// namespace is the leading part of all published metrics.
const namespace = "qed"

// subsystem associated with metrics for server
const subsystem = "server"

type serverMetrics struct {
	Instances prometheus.Gauge
}

func newServerMetrics() *serverMetrics {
	return &serverMetrics{
		Instances: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "instances",
				Help: "Number of QED servers currently running",
			},
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *serverMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.Instances,
	}
}
