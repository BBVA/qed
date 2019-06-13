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

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/util"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
)

var (
	QedPublisherInstancesCount = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "qed_publisher_instances_count",
			Help: "Number of publisher agents running.",
		},
	)

	QedPublisherBatchesReceivedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "qed_publisher_batches_received_total",
			Help: "Number of batches received by publishers.",
		},
	)

	QedPublisherBatchesProcessSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name: "qed_publisher_batches_process_seconds",
			Help: "Duration of Publisher batch processing",
		},
	)
)

var agentPublisherCmd *cobra.Command = &cobra.Command{
	Use:   "publisher",
	Short: "Provides access to the QED gossip publisher agent",
	Long: `Start a QED publisher which process gossip messages sending batch
messages contents to the snapshot storage.`,
	RunE: runAgentPublisher,
}

var agentPublisherCtx context.Context

func init() {
	agentPublisherCtx = configPublisher()
	agentPublisherCmd.MarkFlagRequired("notifier-servers")
	agentPublisherCmd.MarkFlagRequired("store-servers")
	agentCmd.AddCommand(agentPublisherCmd)
}

type publisherConfig struct {
	Notifier *gossip.SimpleNotifierConfig
	Store    *gossip.RestSnapshotStoreConfig
	Tasks    *gossip.SimpleTasksManagerConfig
}

func newPublisherConfig() *publisherConfig {
	return &publisherConfig{
		Notifier: gossip.DefaultSimpleNotifierConfig(),
		Store:    gossip.DefaultRestSnapshotStoreConfig(),
		Tasks:    gossip.DefaultSimpleTasksManagerConfig(),
	}
}

func configPublisher() context.Context {
	conf := newPublisherConfig()
	err := gpflag.ParseTo(conf, agentPublisherCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	ctx := context.WithValue(agentCtx, k("publisher.config"), conf)

	return ctx
}

func runAgentPublisher(cmd *cobra.Command, args []string) error {
	agentConfig := agentCtx.Value(k("agent.config")).(*gossip.Config)
	conf := agentPublisherCtx.Value(k("publisher.config")).(*publisherConfig)

	log.SetLogger("publisher", agentConfig.Log)

	// URL parse
	err := checkPublisherParams(conf)
	if err != nil {
		return err
	}

	notifier := gossip.NewSimpleNotifierFromConfig(conf.Notifier)
	tm := gossip.NewSimpleTasksManagerFromConfig(conf.Tasks)
	store := gossip.NewRestSnapshotStoreFromConfig(conf.Store)

	agent, err := gossip.NewDefaultAgent(agentConfig, nil, store, tm, notifier)
	if err != nil {
		return err
	}

	bp := gossip.NewBatchProcessor(agent, []gossip.TaskFactory{gossip.PrinterFactory{}, publisherFactory{}})
	agent.In.Subscribe(gossip.BatchMessageType, bp, 255)
	defer bp.Stop()

	agent.Start()
	util.AwaitTermSignal(agent.Shutdown)
	return nil
}

func checkPublisherParams(conf *publisherConfig) error {
	var err error
	err = urlParse(conf.Notifier.Endpoint...)
	if err != nil {
		return err
	}

	err = urlParse(conf.Store.Endpoint...)
	if err != nil {
		return err
	}
	return nil
}

type publisherFactory struct {
}

func (p publisherFactory) Metrics() []prometheus.Collector {
	QedPublisherInstancesCount.Inc()
	return []prometheus.Collector{
		QedPublisherInstancesCount,
		QedPublisherBatchesReceivedTotal,
		QedPublisherBatchesProcessSeconds,
	}
}

var errorNoSnapshots error = fmt.Errorf("No snapshots were found on this batch!!")

func (p publisherFactory) New(ctx context.Context) gossip.Task {
	QedPublisherBatchesReceivedTotal.Inc()
	fmt.Println("PublisherFactory creating new Task!")
	a := ctx.Value("agent").(*gossip.Agent)
	b := ctx.Value("batch").(*protocol.BatchSnapshots)

	return func() error {
		timer := prometheus.NewTimer(QedPublisherBatchesProcessSeconds)
		defer timer.ObserveDuration()

		batch := new(protocol.BatchSnapshots)
		batch.Snapshots = make([]*protocol.SignedSnapshot, 0)
		for _, signedSnap := range b.Snapshots {
			_, err := a.Cache.Get(signedSnap.Signature)
			if err != nil {
				log.Debugf("PublishingTask: add snapshot to be published")
				_ = a.Cache.Set(signedSnap.Signature, []byte{0x0}, 0)
				batch.Snapshots = append(batch.Snapshots, signedSnap)
			}
		}

		if len(batch.Snapshots) < 1 {
			return errorNoSnapshots
		}
		log.Debugf("Sending batch to snapshot store: %+v", batch)
		return a.SnapshotStore.PutBatch(batch)
	}
}
