/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, n.A.
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
	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/gossip/monitor"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/util"
	"github.com/spf13/cobra"
)

func newAgentMonitorCommand(ctx *cmdContext, config *gossip.Config) *cobra.Command {

	monitorConfig := monitor.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start a QED monitor",
		Long: `Start a QED monitor that reacts to snapshot batches 
		propagated by QED servers and periodically executes incremental 
		queries to verify the consistency between snaphots`,
		Run: func(cmd *cobra.Command, args []string) {
			log.SetLogger("QedMonitor", ctx.logLevel)

			config.Role = member.Monitor
			monitorConfig.APIKey = ctx.apiKey

			monitor, err := monitor.NewMonitor(*monitorConfig)
			if err != nil {
				log.Fatalf("Failed to start the QED monitor: %v", err)
			}

			agent, err := gossip.NewAgent(config, []gossip.Processor{monitor})
			if err != nil {
				log.Fatalf("Failed to start the QED monitor: %v", err)
			}
			defer agent.Shutdown()

			contacted, err := agent.Join(config.StartJoin)
			if err != nil {
				log.Fatalf("Failed to join the cluster: %v", err)
			}

			log.Debugf("Number of nodes contacted: %d", contacted)

			util.AwaitTermSignal(agent.Leave)
		},
	}

	cmd.Flags().StringSliceVarP(&monitorConfig.QedUrls, "qedUrls", "", []string{}, "Comma-delimited list of QED servers ([host]:port), through which a monitor can make queries")
	cmd.Flags().StringSliceVarP(&monitorConfig.PubUrls, "pubUrls", "", []string{}, "Comma-delimited list of QED servers ([host]:port), through which an monitor can publish alerts")
	cmd.MarkFlagRequired("qedUrls")
	cmd.MarkFlagRequired("pubUrls")

	return cmd
}
