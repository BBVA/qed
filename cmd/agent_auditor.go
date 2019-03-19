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
	v "github.com/spf13/viper"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/gossip/auditor"
	"github.com/bbva/qed/gossip/member"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/util"
)

func newAgentAuditorCommand(ctx *cmdContext, config gossip.Config, agentPreRun func(gossip.Config) gossip.Config) *cobra.Command {

	auditorConfig := auditor.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "auditor",
		Short: "Start a QED auditor",
		Long:  `Start a QED auditor that reacts to snapshot batches propagated by QED servers and periodically executes membership queries to verify the inclusion of events`,
		PreRun: func(cmd *cobra.Command, args []string) {

			log.SetLogger("QEDAuditor", ctx.logLevel)

			// WARN: PersitentPreRun can't be nested and we're using it in cmd/root so inbetween preRuns
			// must be curried.
			config = agentPreRun(config)

			// Bindings
			auditorConfig.MetricsAddr = config.BindAddr // TODO: make MetricsAddr configurable
			auditorConfig.QEDUrls = v.GetStringSlice("agent.server_urls")
			auditorConfig.PubUrls = v.GetStringSlice("agent.alert_urls")

			markSliceStringRequired(auditorConfig.QEDUrls, "qedUrls")
			markSliceStringRequired(auditorConfig.PubUrls, "pubUrls")
		},
		Run: func(cmd *cobra.Command, args []string) {

			config.Role = member.Auditor
			auditorConfig.APIKey = ctx.apiKey

			auditor, err := auditor.NewAuditor(*auditorConfig)
			if err != nil {
				log.Fatalf("Failed to start the QED monitor: %v", err)
			}

			agent, err := gossip.NewAgent(&config, []gossip.Processor{auditor})
			if err != nil {
				log.Fatalf("Failed to start the QED auditor: %v", err)
			}

			contacted, err := agent.Join(config.StartJoin)
			if err != nil {
				log.Fatalf("Failed to join the cluster: %v", err)
			}
			log.Debugf("Number of nodes contacted: %d (%v)", contacted, config.StartJoin)

			defer agent.Shutdown()
			util.AwaitTermSignal(agent.Leave)
		},
	}

	f := cmd.Flags()
	f.StringSliceVarP(&auditorConfig.QEDUrls, "qedUrls", "", []string{}, "Comma-delimited list of QED servers ([host]:port), through which an auditor can make queries")
	f.StringSliceVarP(&auditorConfig.PubUrls, "pubUrls", "", []string{}, "Comma-delimited list of QED servers ([host]:port), through which an auditor can make queries")

	// Lookups
	v.BindPFlag("agent.server_urls", f.Lookup("qedUrls"))
	v.BindPFlag("agent.alert_urls", f.Lookup("pubUrls"))

	return cmd
}
