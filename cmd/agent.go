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

package cmd

import (
	"github.com/bbva/qed/gossip"
	"github.com/spf13/cobra"
)

func newAgentCommand(cmdCtx *cmdContext) *cobra.Command {

	config := gossip.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start a gossip agent for the verifiable log QED",
		Long:  ``,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.EnableCompression = true
		},
		TraverseChildren: true,
	}

	cmd.PersistentFlags().StringVar(&config.NodeName, "node", "", "Unique name for node. If not set, fallback to hostname")
	cmd.PersistentFlags().StringVar(&config.BindAddr, "bind", "", "Bind address for TCP/UDP gossip on (host:port)")
	cmd.PersistentFlags().StringVar(&config.AdvertiseAddr, "advertise", "", "Address to advertise to cluster")
	cmd.PersistentFlags().StringSliceVar(&config.StartJoin, "join", []string{}, "Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")
	cmd.Flags().StringSliceVar(&config.AlertsUrls, "alertsUrls", []string{}, "Comma-delimited list of Alert servers ([host]:port), through which an agent can post alerts")

	cmd.MarkPersistentFlagRequired("node")
	cmd.MarkPersistentFlagRequired("bind")
	cmd.MarkPersistentFlagRequired("join")
	cmd.MarkFlagRequired("alertUrls")

	cmd.AddCommand(newAgentMonitorCommand(cmdCtx, config))
	cmd.AddCommand(newAgentAuditorCommand(cmdCtx, config))
	cmd.AddCommand(newAgentPublisherCommand(cmdCtx, config))

	return cmd

}
