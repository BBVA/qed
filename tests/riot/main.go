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

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
)

type Config struct {
	// flags
	Endpoint         string
	APIKey           string
	Insecure         bool
	WantAdd          bool
	WantIncremental  bool
	WantMembership   bool
	Offload          bool
	Charts           bool
	Profiling        bool
	IncrementalDelta uint
	Offset           uint
	NumRequests      uint
	MaxGoRoutines    uint
	ClusterSize      uint
}

func main() {
	if err := newRiotCommand().Execute(); err != nil {
		os.Exit(-1)
	}
}

func newRiotCommand() *cobra.Command {
	var APIMode bool
	config := Config{}

	cmd := &cobra.Command{
		Use:   "riot",
		Short: "Stresser tool for qed server",
		PreRun: func(cmd *cobra.Command, args []string) {
			config.ClusterSize = uint(viper.GetInt("cluster_size"))
			if config.ClusterSize != 0 && config.ClusterSize != 2 && config.ClusterSize != 4 {
				log.Fatalf("invalid cluster-size specified: %d (only 2 or 4)", config.ClusterSize)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if APIMode {
				Serve(config)
			} else {
				Run(config)
			}
		},
	}

	f := cmd.Flags()
	f.BoolVar(&APIMode, "api", false, "Raise a HTTP api in port 11111 ")
	f.StringVar(&config.Endpoint, "endpoint", "http://localhost:8800", "The endopoint to make the load")
	f.StringVar(&config.APIKey, "apikey", "my-key", "The key to use qed servers")
	f.BoolVar(&config.Insecure, "insecure", false, "Allow self-signed TLS certificates")
	f.BoolVar(&config.WantAdd, "add", false, "Execute add benchmark")
	f.BoolVarP(&config.WantMembership, "membership", "m", false, "Benchmark MembershipProof")
	f.BoolVar(&config.WantIncremental, "incremental", false, "Execute Incremental benchmark")
	f.BoolVar(&config.Offload, "offload", false, "Perform reads only on %50 of the cluster size (With cluster size 2 reads will be performed only on follower1)")
	f.BoolVar(&config.Charts, "charts", false, "Create charts while executing the benchmarks. Output: graph-$testname.png")
	f.BoolVar(&config.Profiling, "profiling", false, "Enable Go profiling with pprof tool. $ go tool pprof -http : http://localhost:6061 ")
	f.UintVarP(&config.IncrementalDelta, "delta", "d", 1000, "Specify delta for the IncrementalProof")
	f.UintVar(&config.NumRequests, "n", 10e4, "Number of requests for the attack")
	f.UintVar(&config.MaxGoRoutines, "r", 10, "Set the concurrency value")
	f.UintVar(&config.Offset, "offset", 0, "The starting version from which we start the load")
	f.UintVar(&config.ClusterSize, "cluster-size", 0, "")

	_ = viper.BindPFlag("cluster_size", f.Lookup("cluster-size"))
	_ = viper.BindEnv("cluster_size", "CLUSTER_SIZE")

	return cmd
}

func Serve(defaultConf Config) {
	mux := http.NewServeMux()
	mux.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}

		var newConf Config
		err := json.NewDecoder(r.Body).Decode(&newConf)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var localConf Config
		if err := mergo.Merge(&localConf, defaultConf); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := mergo.Merge(&localConf, newConf); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf(">>>>>>>>>>>>> %+v", localConf)
	})

	api := &http.Server{
		Addr:    ":18800",
		Handler: mux,
	}

	if err := api.ListenAndServe(); err != http.ErrServerClosed {
		log.Errorf("Can't start Riot API HTTP server: %s", err)
	}
}

func Run(conf Config) {
	var attack Attack

	if conf.WantAdd { // nolint:gocritic
		log.Info("Benchmark ADD")
		attack = Attack{
			kind: "add",
		}
	} else if conf.WantMembership {
		log.Info("Benchmark MEMBERSHIP")

		attack = Attack{
			kind:           "membership",
			balloonVersion: uint64(conf.NumRequests + conf.Offset - 1),
		}
	} else if conf.WantIncremental {
		log.Info("Benchmark INCREMENTAL")

		attack = Attack{
			kind: "incremental",
		}
	}

	attack.config = conf

	attack.Run()
}

type Attack struct {
	kind           string
	balloonVersion uint64

	config  Config
	client  *client.HTTPClient
	reqChan chan uint
	senChan chan Task
}

type Task struct {
	kind string

	event               string
	key                 []byte
	version, start, end uint64
}

func (a *Attack) Run() {
	a.CreateFanOut()
	a.CreateFanIn()

	for i := a.config.Offset; i < a.config.Offset+a.config.NumRequests; i++ {
		a.reqChan <- i
	}

}
func (a *Attack) Shutdown() {
	close(a.reqChan)
	close(a.senChan)
}

func (a *Attack) CreateFanIn() {
	a.reqChan = make(chan uint, a.config.NumRequests/100)

	for rID := uint(0); rID < a.config.MaxGoRoutines; rID++ {
		go func(rID uint) {
			for {
				id, ok := <-a.reqChan
				if !ok {
					log.Infof("Closing mux chan #%d", rID)
					return
				}
				switch a.kind {
				case "add":
					a.senChan <- Task{
						kind:  a.kind,
						event: fmt.Sprintf("event %d", id),
					}
				case "membership":
					a.senChan <- Task{
						kind:    a.kind,
						key:     []byte(fmt.Sprintf("event %d", id)),
						version: a.balloonVersion,
					}

				case "incremental":
					a.senChan <- Task{
						kind:  a.kind,
						start: uint64(id),
						end:   uint64(id + a.config.IncrementalDelta),
					}
				}
			}
		}(rID)
	}

}

func (a *Attack) CreateFanOut() {

	cConf := client.DefaultConfig()
	cConf.Endpoint = a.config.Endpoint
	cConf.APIKey = a.config.APIKey
	cConf.Insecure = a.config.Insecure
	a.client = client.NewHTTPClient(*cConf)
	if err := a.client.Ping(); err != nil {
		panic(err)
	}
	a.senChan = make(chan Task, a.config.NumRequests/100)

	for rID := uint(0); rID < a.config.MaxGoRoutines; rID++ {

		go func(rID uint) {
			for {
				task, ok := <-a.senChan
				if !ok {
					log.Infof("Closing demux chan #%d", rID)
					return
				}

				switch task.kind {
				case "add":
					_, _ = a.client.Add(task.event)
				case "membership":
					_, _ = a.client.Membership(task.key, task.version)
				case "incremental":
					_, _ = a.client.Incremental(task.start, task.end)
				}
			}
		}(rID)
	}
}
