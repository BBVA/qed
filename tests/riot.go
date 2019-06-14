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

package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/imdario/mergo"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/api/metricshttp"
	"github.com/bbva/qed/client"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/util"
)

const riotHelp = `---
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

	RiotEventAdd = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "riot_event_add",
			Help: "Number of events added into the cluster.",
		},
	)
	RiotEventAddFail = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "riot_event_add_fail",
			Help: "Number of events failed to add.",
		},
	)
	RiotQueryMembership = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "riot_query_membership",
			Help: "Number of single events directly verified into the cluster.",
		},
	)
	RiotQueryIncremental = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "riot_query_incremental",
			Help: "Number of range of verified events queried into the cluster.",
		},
	)
	metricsList = []prometheus.Collector{
		RiotEventAdd,
		RiotEventAddFail,
		RiotQueryMembership,
		RiotQueryIncremental,
	}

	registerMetrics sync.Once
)

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

type Riot struct {
	Config Config

	httpServer         *http.Server
	metricsServer      *http.Server
	prometheusRegistry *prometheus.Registry
}

type Config struct {
	// general conf
	Endpoint []string
	APIKey   string
	Insecure bool

	// stress conf
	Kind             string
	Offload          bool
	Profiling        bool
	IncrementalDelta uint
	Offset           uint
	BulkSize         uint
	NumRequests      uint
	MaxGoRoutines    uint
	ClusterSize      uint
}

type Plan [][]Config

type kind string

const (
	add         kind = "add"
	bulk        kind = "bulk"
	membership  kind = "membership"
	incremental kind = "incremental"
)

type Attack struct {
	kind           kind
	balloonVersion uint64

	config  Config
	client  *client.HTTPClient
	senChan chan Task
}

type Task struct {
	kind kind

	events              []string
	version, start, end uint64
}

func main() {
	if err := newRiotCommand().Execute(); err != nil {
		os.Exit(-1)
	}
}

func newRiotCommand() *cobra.Command {
	// Input storage.
	var logLevel string
	var APIMode bool
	riot := Riot{}

	cmd := &cobra.Command{
		Use:   "riot",
		Short: "Stresser tool for qed server",
		Long:  riotHelp,
		PreRun: func(cmd *cobra.Command, args []string) {

			log.SetLogger("Riot", logLevel)

			if riot.Config.Profiling {
				go func() {
					log.Info("	* Starting Riot Profiling server at :6060")
					log.Info(http.ListenAndServe(":6060", nil))
				}()
			}

			if !APIMode && riot.Config.Kind == "" {
				log.Fatal("Argument `kind` is required")
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			riot.Start(APIMode)
		},
	}

	f := cmd.Flags()

	f.StringVarP(&logLevel, "log", "l", "debug", "Choose between log levels: silent, error, info and debug")
	f.BoolVar(&APIMode, "api", false, "Raise a HTTP api in port 7700")

	f.StringSliceVarP(&riot.Config.Endpoint, "endpoint", "e", []string{"127.0.0.1:8800"}, "The endopoint to make the load")
	f.StringVarP(&riot.Config.APIKey, "apikey", "k", "my-key", "The key to use qed servers")
	f.BoolVar(&riot.Config.Insecure, "insecure", false, "Allow self-signed TLS certificates")

	f.StringVar(&riot.Config.Kind, "kind", "", "The kind of load to execute")

	f.BoolVar(&riot.Config.Profiling, "profiling", false, "Enable Go profiling $ go tool pprof")
	f.UintVarP(&riot.Config.IncrementalDelta, "delta", "d", 1000, "Specify delta for the IncrementalProof")
	f.UintVar(&riot.Config.NumRequests, "n", 10e4, "Number of requests for the attack")
	f.UintVar(&riot.Config.MaxGoRoutines, "r", 10, "Set the concurrency value")
	f.UintVar(&riot.Config.Offset, "offset", 0, "The starting version from which we start the load")
	f.UintVar(&riot.Config.BulkSize, "bulk-size", 20, "the size of the bulk in bulk loads (kind: bulk)")

	return cmd
}

func (riot *Riot) Start(APIMode bool) {
	if APIMode {
		riot.Serve()
	} else {
		riot.RunOnce()
	}

	util.AwaitTermSignal(riot.Stop)

	log.Debug("Stopping riot, about to exit...")
}

func (riot *Riot) RunOnce() {
	newAttack(riot.Config)
}

func (riot *Riot) MergeConf(newConf Config) Config {
	conf := riot.Config
	_ = mergo.Merge(&conf, newConf)
	return conf
}

func (riot *Riot) Serve() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, riotHelp)
	})
	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		var err error
		w, r, err = apihttp.PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		var newConf Config
		err = json.NewDecoder(r.Body).Decode(&newConf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newAttack(riot.MergeConf(newConf))
	})

	mux.HandleFunc("/plan", func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		var err error
		w, r, err = apihttp.PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		var plan Plan
		err = json.NewDecoder(r.Body).Decode(&plan)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, batch := range plan {
			for _, conf := range batch {
				wg.Add(1)
				go func(conf Config) {
					newAttack(riot.MergeConf(conf))
					wg.Done()
				}(conf)

			}
			wg.Wait()
		}
	})

	// Metrics server
	r := prometheus.NewRegistry()
	Register(r)
	riot.prometheusRegistry = r
	metricsMux := metricshttp.NewMetricsHTTP(r)
	log.Debug("	* Starting Riot Metrics server at :17700")
	riot.metricsServer = &http.Server{Addr: ":17700", Handler: metricsMux}

	go func() {
		if err := riot.metricsServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start metrics HTTP server: %s", err)
		}
	}()

	// API server
	riot.httpServer = &http.Server{Addr: ":7700", Handler: mux}
	log.Debug("	* Starting Riot HTTP server at :7700")
	go func() {
		if err := riot.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("Can't start Riot API HTTP server: %s", err)
		}
	}()
}

func (riot *Riot) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Debug("Stopping metrics server...")
	err := riot.metricsServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	log.Debug("Stopping HTTP server...")
	err = riot.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}

	return nil
}

func newAttack(conf Config) {

	// QED client
	transport := http.DefaultTransport.(*http.Transport)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: conf.Insecure}
	httpClient := http.DefaultClient
	httpClient.Transport = transport
	client, err := client.NewHTTPClient(
		client.SetHttpClient(httpClient),
		client.SetURLs(conf.Endpoint[0], conf.Endpoint[1:]...),
		client.SetAPIKey(conf.APIKey),
		client.SetReadPreference(client.Any),
		client.SetMaxRetries(1),
		client.SetTopologyDiscovery(true),
		client.SetHealthChecks(true),
		client.SetHealthCheckTimeout(2*time.Second),   // default value
		client.SetHealthCheckInterval(60*time.Second), // default value
		client.SetAttemptToReviveEndpoints(true),
		client.SetHasherFunction(hashing.NewSha256Hasher),
	)

	if err != nil {
		panic(err)
	}

	attack := Attack{
		client:         client,
		config:         conf,
		kind:           kind(conf.Kind),
		balloonVersion: uint64(conf.NumRequests + conf.Offset - 1),
	}

	if err := attack.client.Ping(); err != nil {
		panic(err)
	}

	attack.Run()
}

func (a *Attack) Run() {
	var wg sync.WaitGroup
	a.senChan = make(chan Task)

	for rID := uint(0); rID < a.config.MaxGoRoutines; rID++ {
		wg.Add(1)
		go func(rID uint) {
			for {
				task, ok := <-a.senChan
				if !ok {
					log.Debugf("!!! close: %d", rID)
					wg.Done()
					return
				}

				switch task.kind {
				case add:
					log.Debugf("Adding: %s", task.events[0])
					_, err := a.client.Add(task.events[0])
					if err != nil {
						RiotEventAddFail.Inc()
						log.Debugf("Error adding event: version %d. Error: %s", task.version, err)
					} else {
						RiotEventAdd.Inc()
					}
				case bulk:
					bulkSize := len(task.events)
					log.Debugf("Inserting bulk: version %d, size %d, first event: %s", task.version, bulkSize, task.events[0])
					_, err := a.client.AddBulk(task.events)
					if err != nil {
						RiotEventAddFail.Add(float64(bulkSize))
						log.Debugf("Error inserting bulk: version %d, size %d. Error: %s", task.version, bulkSize, err)
					} else {
						RiotEventAdd.Add(float64(bulkSize))
					}
				case membership:
					log.Debugf("Querying membership: event %s", task.events[0])
					_, _ = a.client.Membership([]byte(task.events[0]), &task.version)
					RiotQueryMembership.Inc()
				case incremental:
					log.Debugf("Querying incremental: start %d, end %d", task.start, task.end)
					_, _ = a.client.Incremental(task.start, task.end)
					RiotQueryIncremental.Inc()
				}
			}
		}(rID)
	}

	hasReqs := func(i uint) bool {
		return i < a.config.Offset+a.config.NumRequests
	}

	hasBulk := func(j, i uint) bool {
		return i < j+a.config.BulkSize && hasReqs(i)
	}

	for i := a.config.Offset; hasReqs(i); i++ {
		task := Task{
			kind:    a.kind,
			events:  []string{},
			version: a.balloonVersion,
			start:   uint64(i),
			end:     uint64(i + a.config.IncrementalDelta),
		}

		for j := i; hasBulk(j, i); i++ {
			task.events = append(task.events, fmt.Sprintf("event %d", i))
		}

		a.senChan <- task
	}

	close(a.senChan)
	wg.Wait()
}
