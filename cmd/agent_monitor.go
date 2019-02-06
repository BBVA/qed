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
	"github.com/spf13/cobra"
	v "github.com/spf13/viper"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/gossip/monitor"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/util"
)

func newAgentMonitorCommand(ctx cmdContext, config gossip.Config, agentPreRun func(gossip.Config) gossip.Config) *cobra.Command {

	monitorConfig := monitor.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Start a QED monitor",
		Long:  `Start a QED monitor that reacts to snapshot batches propagated by QED servers and periodically executes incremental queries to verify the consistency between snaphots`,
		PreRun: func(cmd *cobra.Command, args []string) {

			log.SetLogger("QEDMonitor", ctx.logLevel)

			// WARN: PersitentPreRun can't be nested and we're using it in cmd/root so inbetween preRuns
			// must be curried.
			config = agentPreRun(config)

			// Bindings
			monitorConfig.QEDUrls = v.GetStringSlice("agent.server_urls")
			monitorConfig.PubUrls = v.GetStringSlice("agent.alert_urls")
			markSliceStringRequired(monitorConfig.QEDUrls, "qedUrls")
			markSliceStringRequired(monitorConfig.PubUrls, "pubUrls")

		},
		Run: func(cmd *cobra.Command, args []string) {

			config.Role = member.Monitor
			monitorConfig.APIKey = ctx.apiKey

			monitor, err := monitor.NewMonitor(*monitorConfig)
			if err != nil {
				log.Fatalf("Failed to start the QED monitor: %v", err)
			}

			agent, err := gossip.NewAgent(&config, []gossip.Processor{monitor})
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

	f := cmd.Flags()
	f.StringSliceVarP(&monitorConfig.QEDUrls, "qedUrls", "", []string{}, "Comma-delimited list of QED servers ([host]:port), through which a monitor can make queries")
	f.StringSliceVarP(&monitorConfig.PubUrls, "pubUrls", "", []string{}, "Comma-delimited list of QED servers ([host]:port), through which an monitor can publish alerts")

	// Lookups
	v.BindPFlag("agent.server_urls", f.Lookup("qedUrls"))
	v.BindPFlag("agent.alert_urls", f.Lookup("pubUrls"))

	return cmd
}
