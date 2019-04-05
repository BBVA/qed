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

package history

import "github.com/prometheus/client_golang/prometheus"

const namespace = "qed"
const subSystem = "history"

var (
	AddTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "add_total",
			Help:      "Number of the events added to the history tree.",
		},
	)
	MembershipTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "membership_total",
			Help:      "Number of membership queries",
		},
	)
	ConsistencyTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subSystem,
			Name:      "consistency_total",
			Help:      "Number of consistency queries",
		},
	)
)
