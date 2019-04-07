/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, n.A.
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

package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/util"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var (
	QedMonitorInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_monitor_instances_count",
			Help: "Number of monitor agents running.",
		},
	)

	QedMonitorBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_monitor_batches_received_total",
			Help: "Number of batches received by monitors.",
		},
	)

	QedMonitorBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_monitor_batches_process_seconds",
			Help: "Duration of Monitor batch processing",
		},
	)

	QedMonitorGetIncrementalProofErrTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_monitor_get_incremental_proof_err_total",
			Help: "Number of errors trying to get incremental proofs by monitors.",
		},
	)
)

var agentMonitorCmd *cobra.Command = &cobra.Command{
	Use:              "monitor",
	Short:            "Provides access to the QED gossip monitor agent",
	TraverseChildren: true,
	RunE:             runAgentMonitor,
}

var agentMonitorCtx context.Context

func init() {
	agentMonitorCtx = configMonitor()
	agentCmd.AddCommand(agentMonitorCmd)
}

type monitorConfig struct {
	Qed      *client.Config
	Notifier *gossip.DefaultNotifierConfig
	Store    *gossip.RestSnapshotStoreConfig
	Tasks    *gossip.DefaultTasksManagerConfig
}

func newMonitorConfig() *monitorConfig {
	return &monitorConfig{
		Qed: client.DefaultConfig(),
		Notifier: &gossip.DefaultNotifierConfig{
			DialTimeout:  200 * time.Millisecond,
			QueueTimeout: 100 * time.Millisecond,
			ReadTimeout:  200 * time.Millisecond,
		},
		Store: &gossip.RestSnapshotStoreConfig{
			DialTimeout:  200 * time.Millisecond,
			QueueTimeout: 100 * time.Millisecond,
			ReadTimeout:  200 * time.Millisecond,
		},
		Tasks: &gossip.DefaultTasksManagerConfig{
			QueueTimeout: 100 * time.Millisecond,
			Interval:     200 * time.Millisecond,
			MaxTasks:     10,
		},
	}
}

func configMonitor() context.Context {
	conf := newMonitorConfig()
	err := gpflag.ParseTo(conf, agentMonitorCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	ctx := context.WithValue(agentCtx, k("monitor.config"), conf)

	return ctx
}

func runAgentMonitor(cmd *cobra.Command, args []string) error {
	agentConfig := agentMonitorCtx.Value(k("agent.config")).(*gossip.Config)
	conf := agentMonitorCtx.Value(k("monitor.config")).(*monitorConfig)

	log.SetLogger("monitor", agentConfig.Log)

	notifier := gossip.NewDefaultNotifierFromConfig(conf.Notifier)
	qed, err := client.NewHTTPClientFromConfig(conf.Qed)
	if err != nil {
		return err
	}
	tm := gossip.NewDefaultTasksManagerFromConfig(conf.Tasks)
	store := gossip.NewRestSnapshotStoreFromConfig(conf.Store)

	agent, err := gossip.NewDefaultAgent(agentConfig, qed, store, tm, notifier)
	if err != nil {
		return err
	}

	bp := gossip.NewBatchProcessor(agent, []gossip.TaskFactory{gossip.PrinterFactory{}, incrementalFactory{}})
	agent.In.Subscribe(gossip.BatchMessageType, bp, 255)
	defer bp.Stop()

	agent.Start()
	util.AwaitTermSignal(agent.Shutdown)
	return nil
}

type incrementalFactory struct{}

func (i incrementalFactory) Metrics() []prometheus.Collector {
	return []prometheus.Collector{
		QedMonitorInstancesCount,
		QedMonitorBatchesReceivedTotal,
		QedMonitorBatchesProcessSeconds,
		QedMonitorGetIncrementalProofErrTotal,
	}
}

func (i incrementalFactory) New(ctx context.Context) gossip.Task {
	a := ctx.Value("agent").(*gossip.Agent)
	b := ctx.Value("batch").(*protocol.BatchSnapshots)

	return func() error {
		timer := prometheus.NewTimer(QedMonitorBatchesProcessSeconds)
		defer timer.ObserveDuration()

		first := b.Snapshots[0].Snapshot
		last := b.Snapshots[len(b.Snapshots)-1].Snapshot

		resp, err := a.Qed.Incremental(first.Version, last.Version)
		if err != nil {
			QedMonitorGetIncrementalProofErrTotal.Inc()
			log.Infof("Monitor is unable to get incremental proof from QED server: %s", err.Error())
			return err
		}
		ok := a.Qed.VerifyIncremental(resp, first, last, hashing.NewSha256Hasher())
		if !ok {
			a.Notifier.Alert(fmt.Sprintf("Monitor is unable to verify incremental proof from %d to %d", first.Version, last.Version))
			log.Infof("Monitor is unable to verify incremental proof from %d to %d", first.Version, last.Version)
		}
		log.Debugf("Monitor verified a consistency proof between versions %d and %d: %v\n", first.Version, last.Version, ok)
		return nil
	}
}

