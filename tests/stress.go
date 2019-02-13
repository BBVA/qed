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

type Config struct {
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
	config := Config{}

	cmd := &cobra.Command{
		Use:   "riot",
		Short: "Stresser tool for qed server",
		PreRun: func(cmd *cobra.Command, args []string) {
			config.clusterSize = uint(viper.GetInt("cluster_size"))
			if config.clusterSize != 0 && config.clusterSize != 2 && config.clusterSize != 4 {
				log.Fatalf("invalid cluster-size specified: %d (only 2 or 4)", config.clusterSize)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(config)
		},
	}

	f := cmd.Flags()
	f.StringVar(&config.endpoint, "endpoint", "http://localhost:8800", "The endopoint to make the load")
	f.StringVar(&config.apiKey, "apikey", "my-key", "The key to use qed servers")
	f.BoolVar(&config.insecure, "insecure", false, "Allow self-signed TLS certificates")
	f.BoolVar(&config.wantAdd, "add", false, "Execute add benchmark")
	f.BoolVarP(&config.wantMembership, "membership", "m", false, "Benchmark MembershipProof")
	f.BoolVar(&config.wantIncremental, "incremental", false, "Execute Incremental benchmark")
	f.BoolVar(&config.offload, "offload", false, "Perform reads only on %50 of the cluster size (With cluster size 2 reads will be performed only on follower1)")
	f.BoolVar(&config.charts, "charts", false, "Create charts while executing the benchmarks. Output: graph-$testname.png")
	f.BoolVar(&config.profiling, "profiling", false, "Enable Go profiling with pprof tool. $ go tool pprof -http : http://localhost:6061 ")
	f.UintVarP(&config.incrementalDelta, "delta", "d", 1000, "Specify delta for the IncrementalProof")
	f.UintVar(&config.numRequests, "n", 10e4, "Number of requests for the attack")
	f.UintVar(&config.maxGoRoutines, "r", 10, "Set the concurrency value")
	f.UintVar(&config.offset, "offset", 0, "The starting version from which we start the load")
	f.UintVar(&config.clusterSize, "cluster-size", 0, "")

	_ = viper.BindPFlag("cluster_size", f.Lookup("cluster-size"))
	_ = viper.BindEnv("cluster_size", "CLUSTER_SIZE")

	return cmd
}

func Run(conf Config) error {
	var attack Attack

	if conf.wantAdd { // nolint:gocritic
		log.Info("Benchmark ADD")
		attack = Attack{
			kind: "add",
		}
	} else if conf.wantMembership {
		log.Info("Benchmark MEMBERSHIP")

		attack = Attack{
			kind:           "membership",
			balloonVersion: uint64(conf.numRequests + conf.offset - 1),
		}
	} else if conf.wantIncremental {
		log.Info("Benchmark INCREMENTAL")

		attack = Attack{
			kind: "incremental",
		}
	}

	attack.config = conf

	attack.Run()
	return nil
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

	for i := a.config.offset; i < a.config.offset+a.config.numRequests; i++ {
		a.reqChan <- i
	}

}
func (a *Attack) Shutdown() {
	close(a.reqChan)
	close(a.senChan)
}

func (a *Attack) CreateFanIn() {
	a.reqChan = make(chan uint, a.config.numRequests/100)

	for rID := uint(0); rID < a.config.maxGoRoutines; rID++ {
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
						end:   uint64(id + a.config.incrementalDelta),
					}
				}
			}
		}(rID)
	}

}

func (a *Attack) CreateFanOut() {

	cConf := client.DefaultConfig()
	cConf.Endpoint = a.config.endpoint
	cConf.APIKey = a.config.apiKey
	cConf.Insecure = a.config.insecure
	a.client = client.NewHTTPClient(*cConf)
	if err := a.client.Ping(); err != nil {
		panic(err)
	}
	a.senChan = make(chan Task, a.config.numRequests/100)

	for rID := uint(0); rID < a.config.maxGoRoutines; rID++ {

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
