/*
   copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   you may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   withouT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   see the License for the specific language governing permissions and
   limitations under the License.
*/

package workload

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const WorkloadHelp = `---

workloader:
	this program runs as a single workloader (default) or as a server to receive
	"plans" (Config structs) through a small web API.

API:
  /run:
    examples:
	  # simple add
	  curl -XPOST -H'Content-type: application/json' http://127.0.0.1:7700/run \
	  -d'{"kind":"add"}'

	  # complex config example
	  curl -XPOST -H'Content-type: application/json' http://127.0.0.1:7700/run \
	  -d'{"kind": "incremental", "insecure":true, "endpoint": "https://qedserver:8800,qedserver1:8801"}'

  /plan:
	examples:
	  # like a simple run
	  curl -XPOST -H'Content-type: application/json' http://127.0.0.1:7700/plan \
	  -d'[[{"kind": "add"}]]'

	  # secuential
	  curl -XPOST -H'Content-type: application/json' http://127.0.0.1:7700/plan \
	  -d'[[{"kind": "add"}], [{"kind": "membership"}]]'

	  # parallel
	  curl -XPOST -H'Content-type: application/json' http://127.0.0.1:7700/plan \
	  -d'[[{"kind": "add"}, {"kind": "membership"}]]'
`

var (
	// Client
	workloadEventAdd = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "workload_event_add",
			Help: "Number of events added into the cluster.",
		},
	)
	workloadEventAddFail = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "workload_event_add_fail",
			Help: "Number of events failed to add.",
		},
	)
	workloadQueryMembership = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "workload_query_membership",
			Help: "Number of single events directly verified into the cluster.",
		},
	)
	workloadQueryIncremental = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "workload_query_incremental",
			Help: "Number of range of verified events queried into the cluster.",
		},
	)
	metricsList = []prometheus.Collector{
		workloadEventAdd,
		workloadEventAddFail,
		workloadQueryMembership,
		workloadQueryIncremental,
	}

	registerMetrics sync.Once
)

type Config struct {
	// general conf
	Endpoints []string `desc:"The endopoint to make the load"`
	Insecure  bool     `desc:"Allow self-signed TLS certificates"`
	Log       string   `desc:"Set log level to info, error or debug"`

	// stress conf
	APIMode          bool   `desc:"Enable API Mode"`
	Kind             string `desc:"The kind of load to execute"`
	Offload          bool   `desc:"Slow down request speed"`
	Profiling        bool   `desc:"Enable Go profiling $ go tool pprof"`
	IncrementalDelta uint   `desc:"Specify delta for the IncrementalProof"`
	Offset           uint   `desc:"The starting version from which we start the load"`
	BulkSize         uint   `desc:"The size of the bulk in bulk loads (kind: bulk)"`
	NumRequests      uint   `desc:"Number of requests for the attack"`
	MaxGoRoutines    uint   `desc:"Set the concurrency value"`
}

func DefaultConfig() *Config {
	return &Config{
		Endpoints:        []string{"http://localhost:8800"},
		APIMode:          false,
		Log:              "info",
		Kind:             "",
		Offload:          false,
		Profiling:        false,
		IncrementalDelta: 1000,
		Offset:           0,
		BulkSize:         20,
		NumRequests:      10e4,
		MaxGoRoutines:    10,
	}
}
