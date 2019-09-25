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
	"os"
	"sync/atomic"
	"time"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/client"
	"github.com/bbva/qed/gossip"
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
	Use:   "monitor",
	Short: "Provides access to the QED gossip monitor agent",
	Long: `Stats a QED monitor which process gossip messages measuring
the lag between the gossip received messages and the contents of the
snapshotsore. It also executes incremental proof verification against
some of the snapshots received.`,
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
	Notifier *gossip.SimpleNotifierConfig
	Store    *gossip.RestSnapshotStoreConfig
	Tasks    *gossip.SimpleTasksManagerConfig
}

func newMonitorConfig() *monitorConfig {
	conf := client.DefaultConfig()
	conf.AttemptToReviveEndpoints = true
	conf.ReadPreference = client.Any
	conf.MaxRetries = 1
	return &monitorConfig{
		Qed:      conf,
		Notifier: gossip.DefaultSimpleNotifierConfig(),
		Store:    gossip.DefaultRestSnapshotStoreConfig(),
		Tasks:    gossip.DefaultSimpleTasksManagerConfig(),
	}
}

func configMonitor() context.Context {
	conf := newMonitorConfig()
	err := gpflag.ParseTo(conf, agentMonitorCmd.PersistentFlags())
	if err != nil {
		fmt.Printf("Cannot parse monitor flags: %v\n", err)
		os.Exit(1)
	}

	ctx := context.WithValue(agentCtx, k("monitor.config"), conf)

	return ctx
}

func runAgentMonitor(cmd *cobra.Command, args []string) error {
	agentConfig := agentMonitorCtx.Value(k("agent.config")).(*gossip.Config)
	conf := agentMonitorCtx.Value(k("monitor.config")).(*monitorConfig)

	// create main logger
	logOpts := &log.LoggerOptions{
		Name:            "qed.monitor",
		IncludeLocation: true,
		Level:           log.LevelFromString(agentConfig.Log),
		Output:          log.DefaultOutput,
		TimeFormat:      log.DefaultTimeFormat,
	}
	logger := log.New(logOpts)

	// URL parse
	err := checkMonitorParams(conf)
	if err != nil {
		return err
	}

	notifier := gossip.NewSimpleNotifierFromConfig(conf.Notifier, logger.Named("agent.notifier"))
	qed, err := client.NewHTTPClientFromConfig(conf.Qed)
	if err != nil {
		return err
	}
	tm := gossip.NewSimpleTasksManagerFromConfig(conf.Tasks, logger.Named("agent.task-manager"))
	store := gossip.NewRestSnapshotStoreFromConfig(conf.Store)

	agent, err := gossip.NewDefaultAgent(agentConfig, qed, store, tm, notifier, logger.Named("agent"))
	if err != nil {
		return err
	}

	lagf := newLagFactory(1*time.Second, logger.Named("agent.lag-factory"))
	lagf.start()
	incF := incrementalFactory{logger.Named("agent.incremental-factory")}
	defer lagf.stop()
	bp := gossip.NewBatchProcessor(agent, []gossip.TaskFactory{incF, lagf}, logger.Named("agent.processor"))
	agent.In.Subscribe(gossip.BatchMessageType, bp, 255)
	defer bp.Stop()

	agent.Start()

	QedMonitorInstancesCount.Inc()

	util.AwaitTermSignal(agent.Shutdown)
	return nil
}

func checkMonitorParams(conf *monitorConfig) error {
	var err error
	err = urlParse(conf.Notifier.Endpoint...)
	if err != nil {
		return err
	}

	err = urlParse(conf.Store.Endpoint...)
	if err != nil {
		return err
	}

	err = urlParse(conf.Qed.Endpoints...)
	if err != nil {
		return err
	}
	return nil
}

type incrementalFactory struct {
	log log.Logger
}

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

		firstSnap := balloon.Snapshot(*b.Snapshots[0].Snapshot)
		lastSnap := balloon.Snapshot(*b.Snapshots[len(b.Snapshots)-1].Snapshot)

		proof, err := a.Qed.Incremental(firstSnap.Version, lastSnap.Version)
		if err != nil {
			QedMonitorGetIncrementalProofErrTotal.Inc()
			_ = a.Notifier.Alert(fmt.Sprintf("Monitor is unable to get incremental proof from QED server: %s", err.Error()))
			i.log.Infof("Monitor is unable to get incremental proof from QED server: %s", err.Error())
			return err
		}

		ok, err := a.Qed.IncrementalVerify(proof, &firstSnap, &lastSnap)
		if err != nil {
			i.log.Infof("Error verifying incremental proof: %v", err)
			return nil
		}
		if !ok {
			_ = a.Notifier.Alert(fmt.Sprintf("Monitor is unable to verify incremental proof from %d to %d", firstSnap.Version, lastSnap.Version))
			i.log.Infof("Monitor is unable to verify incremental proof from %d to %d", firstSnap.Version, lastSnap.Version)
		}
		i.log.Debugf("Monitor verified a consistency proof between versions %d and %d: %v\n", firstSnap.Version, lastSnap.Version, ok)
		return nil
	}
}

type lagFactory struct {
	lastVersion uint64
	rate        uint64
	counter     uint64
	ticker      *time.Ticker
	quit        chan struct{}
	log         log.Logger
}

func newLagFactory(t time.Duration, l log.Logger) *lagFactory {
	logger := l
	if logger == nil {
		logger = log.L()
	}
	return &lagFactory{
		ticker: time.NewTicker(t),
		quit:   make(chan struct{}),
		log:    logger,
	}
}

func (l *lagFactory) stop() {
	close(l.quit)
}

func (l *lagFactory) start() {
	go func() {
		for {
			select {
			case <-l.ticker.C:
				c := atomic.SwapUint64(&l.counter, 0)
				atomic.StoreUint64(&l.rate, c)
			case <-l.quit:
				l.ticker.Stop()
				return
			}
		}
	}()
}

func (l lagFactory) Metrics() []prometheus.Collector {
	return []prometheus.Collector{}
}

func (l *lagFactory) New(ctx context.Context) gossip.Task {
	a := ctx.Value("agent").(*gossip.Agent)
	b := ctx.Value("batch").(*protocol.BatchSnapshots)

	counter := atomic.AddUint64(&l.counter, uint64(len(b.Snapshots)))
	lastVersion := atomic.LoadUint64(&l.lastVersion)

	QedMonitorBatchesReceivedTotal.Inc()

	return func() error {
		timer := prometheus.NewTimer(QedMonitorBatchesProcessSeconds)
		defer timer.ObserveDuration()

		last := b.Snapshots[len(b.Snapshots)-1].Snapshot
		localLag := uint64(0)

		if lastVersion < last.Version {
			localLag = last.Version - lastVersion
			atomic.StoreUint64(&l.lastVersion, last.Version)
		}

		rate := atomic.LoadUint64(&l.rate)

		if localLag > rate {
			l.log.Infof("Gossip lag %d > Rate %d", localLag, rate)
		}

		count, err := a.SnapshotStore.Count()
		if err != nil {
			return err
		}

		storeLag := uint64(0)
		if lastVersion > count {
			storeLag = lastVersion - count
		}

		if storeLag > rate {
			err := a.Notifier.Alert(fmt.Sprintf("Lag between gossip and snapshot store: %d", storeLag))
			if err != nil {
				l.log.Infof("LagTask had an error sending a notification: %v", err)
			}
			l.log.Debugf("Lag between gossip and snapshot store: last seen version %d - store count %d  = %d", lastVersion, count, storeLag)
		}
		l.log.Debugf("Lag status: Rate: %d Counter: %d, Local Lag: %d Store Lag: %d", rate, counter, localLag, storeLag)
		return nil
	}
}
