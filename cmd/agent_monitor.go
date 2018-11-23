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
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/util"
	"github.com/spf13/cobra"
)

func newAgentMonitorCommand(ctx *agentContext) *cobra.Command {

	var qedEndpoints []string

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start a QED monitor",
		Long: `Start a QED monitor that reacts to snapshot batches 
		propagated by QED servers and periodically executes incremental 
		queries to verify the consistency between commitments`,
		Run: func(cmd *cobra.Command, args []string) {

			log.SetLogger("QedMonitor", logLevel)

			config := ctx.config
			//monitorConfig := monitor.DefaultConfig()

			agent, err := gossip.NewAgent(config, []gossip.Processor{gossip.DummyProcessor{}})
			if err != nil {
				log.Fatalf("Failed to start the QED monitor: %v", err)
			}

			contacted, err := agent.Join(config.StartJoin)
			if err != nil {
				log.Fatalf("Failed to join the cluster: %v", err)
			}
			log.Debugf("Number of nodes contacted: %d", contacted)

			agent.Start()
			defer agent.Shutdown()
			util.AwaitTermSignal(agent.Leave)
		},
	}

	cmd.Flags().StringSliceVarP(&qedEndpoints, "endpoints", "", []string{}, "Comma-delimited list of QED servers ([host]:port), through which a monitor can make queries")
	cmd.MarkPersistentFlagRequired("endpoints")

	return cmd
}
