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
	"github.com/spf13/cobra"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/gossip/publisher"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/util"
)

func newAgentPublisherCommand(ctx *cmdContext, config gossip.Config, agentPreRun func(gossip.Config) gossip.Config) *cobra.Command {

	publisherConfig := publisher.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "publisher",
		Short: "Start a QED publisher",
		Long:  `Start a QED publisher that reacts to snapshot batches propagated by QED servers and periodically publishes them to a certain log storage.`,
		PreRun: func(cmd *cobra.Command, args []string) {

			log.SetLogger("QEDPublisher", ctx.logLevel)

			// WARN: PersitentPreRun can't be nested and we're using it in
			// cmd/root so inbetween preRuns must be curried.
			config = agentPreRun(config)

		},
		Run: func(cmd *cobra.Command, args []string) {

			config.Role = member.Publisher
			alertsCh := make(chan string, 100)
			publisher, err := publisher.NewPublisher(*publisherConfig, alertsCh)
			if err != nil {
				log.Fatalf("Failed to start the QED publisher: %v", err)
			}
			metricsServer := metrics.NewServer(config.MetricsAddr)
			agent, err := gossip.NewAgent(&config, []gossip.Processor{publisher}, metricsServer, alertsCh)
			if err != nil {
				log.Fatalf("Failed to start the QED publisher: %v", err)
			}

			contacted, err := agent.Join(config.StartJoin)
			if err != nil {
				log.Fatalf("Failed to join the cluster: %v", err)
			}
			log.Debugf("Number of nodes contacted: %d", contacted)

			defer agent.Shutdown()
			util.AwaitTermSignal(agent.Leave)
		},
	}

	f := cmd.Flags()

	return cmd
}
