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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
)

type Riot struct {
	// flags
	endpoint         string
	apiKey           string
	insecure         bool
	wantAdd          bool
	wantIncremental  bool
	wantMembership   bool
	offload          bool
	charts           bool
	profiling        bool
	incrementalDelta uint
	offset           uint
	numRequests      uint
	maxGoRoutines    uint
	clusterSize      uint
}

func main() {
	if err := newRiotCommand().Execute(); err != nil {
		os.Exit(-1)
	}
}

func newRiotCommand() *cobra.Command {
	riot := Riot{}

	cmd := &cobra.Command{
		Use:   "riot",
		Short: "Stresser tool for qed server",
		PreRun: func(cmd *cobra.Command, args []string) {
			riot.clusterSize = uint(viper.GetInt("cluster_size"))
			if riot.clusterSize != 0 && riot.clusterSize != 2 && riot.clusterSize != 4 {
				log.Fatalf("invalid cluster-size specified: %d (only 2 or 4)", riot.clusterSize)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(riot)
		},
	}

	f := cmd.Flags()
	f.StringVar(&riot.endpoint, "endpoint", "http://localhost:8800", "The endopoint to make the load")
	f.StringVar(&riot.apiKey, "apikey", "my-key", "The key to use qed servers")
	f.BoolVar(&riot.insecure, "insecure", false, "Allow self-signed TLS certificates")
	f.BoolVar(&riot.wantAdd, "add", false, "Execute add benchmark")
	f.BoolVarP(&riot.wantMembership, "membership", "m", false, "Benchmark MembershipProof")
	f.BoolVar(&riot.wantIncremental, "incremental", false, "Execute Incremental benchmark")
	f.BoolVar(&riot.offload, "offload", false, "Perform reads only on %50 of the cluster size (With cluster size 2 reads will be performed only on follower1)")
	f.BoolVar(&riot.charts, "charts", false, "Create charts while executing the benchmarks. Output: graph-$testname.png")
	f.BoolVar(&riot.profiling, "profiling", false, "Enable Go profiling with pprof tool. $ go tool pprof -http : http://localhost:6061 ")
	f.UintVarP(&riot.incrementalDelta, "delta", "d", 1000, "Specify delta for the IncrementalProof")
	f.UintVar(&riot.numRequests, "n", 10e4, "Number of requests for the attack")
	f.UintVar(&riot.maxGoRoutines, "r", 10, "Set the concurrency value")
	f.UintVar(&riot.offset, "offset", 0, "The starting version from which we start the load")
	f.UintVar(&riot.clusterSize, "cluster-size", 0, "")

	_ = viper.BindPFlag("cluster_size", f.Lookup("cluster-size"))
	_ = viper.BindEnv("cluster_size", "CLUSTER_SIZE")

	return cmd
}

func Run(r Riot) error {
	var attack Attack

	if r.wantAdd { // nolint:gocritic
		log.Info("Benchmark ADD")
		attack = Attack{
			kind: "add",
		}
	} else if r.wantMembership {
		log.Info("Benchmark MEMBERSHIP")

		attack = Attack{
			kind:           "membership",
			balloonVersion: uint64(r.numRequests + r.offset - 1),
		}
	} else if r.wantIncremental {
		log.Info("Benchmark INCREMENTAL")

		attack = Attack{
			kind: "incremental",
		}
	}
	attack.riot = r

	attack.Run()
	return nil
}

type Attack struct {
	kind           string
	balloonVersion uint64

	riot   Riot
	client *client.HTTPClient
	ch     chan Task
}

type Task struct {
	kind string

	event               string
	key                 []byte
	version, start, end uint64
}

func (a *Attack) Run() {
	a.CreateFanOut()
	a.FanIn()
}

func (a *Attack) FanIn() {
	reqChan := make(chan uint, a.riot.numRequests)

	for rID := uint(0); rID < a.riot.maxGoRoutines; rID++ {
		go func(rID uint) {
			for {
				id, ok := <-reqChan
				if !ok {
					log.Infof("Closing mux chan #%d", rID)
					return
				}
				switch a.kind {
				case "add":
					a.ch <- Task{
						kind:  a.kind,
						event: fmt.Sprintf("event %d", id),
					}
				case "membership":
					a.ch <- Task{
						kind:    a.kind,
						key:     []byte(fmt.Sprintf("event %d", id)),
						version: a.balloonVersion,
					}

				case "incremental":
					a.ch <- Task{
						kind:  a.kind,
						start: uint64(id),
						end:   uint64(id + a.riot.incrementalDelta),
					}
				}
			}
		}(rID)
	}

	for i := a.riot.offset; i < a.riot.offset+a.riot.numRequests; i++ {
		reqChan <- i
	}

}

func (a *Attack) CreateFanOut() {

	cConf := client.DefaultConfig()
	cConf.Endpoint = a.riot.endpoint
	cConf.APIKey = a.riot.apiKey
	cConf.Insecure = a.riot.insecure
	a.client = client.NewHTTPClient(*cConf)
	if err := a.client.Ping(); err != nil {
		panic(err)
	}
	a.ch = make(chan Task, a.riot.numRequests)

	for rID := uint(0); rID < a.riot.maxGoRoutines; rID++ {

		go func(rID uint) {
			for {
				task, ok := <-a.ch
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

func (a *Attack) Shutdown() {
	close(a.ch)
}
