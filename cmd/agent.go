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
	"regexp"

	"github.com/spf13/cobra"
	v "github.com/spf13/viper"

	"github.com/bbva/qed/gossip"
)

func newAgentCommand(cmdCtx cmdContext, args []string) *cobra.Command {

	config := gossip.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start a gossip agent for the verifiable log QED",
	}

	f := cmd.PersistentFlags()
	f.StringVar(&config.NodeName, "node", "", "Unique name for node. If not set, fallback to hostname")
	f.StringVar(&config.BindAddr, "bind", "", "Bind address for TCP/UDP gossip on (host:port)")
	f.StringVar(&config.AdvertiseAddr, "advertise", "", "Address to advertise to cluster")
	f.StringSliceVar(&config.StartJoin, "join", []string{}, "Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")
	f.StringSliceVar(&config.AlertsUrls, "alertsUrls", []string{}, "Comma-delimited list of Alert servers ([host]:port), through which an agent can post alerts")

	// Lookups
	v.BindPFlag("agent.node", f.Lookup("node"))
	v.BindPFlag("agent.bind", f.Lookup("bind"))
	v.BindPFlag("agent.advertise", f.Lookup("advertise"))
	v.BindPFlag("agent.join", f.Lookup("join"))
	v.BindPFlag("agent.alert_urls", f.Lookup("alertsUrls"))

	agentPreRun := func(config gossip.Config) gossip.Config {
		config.EnableCompression = true
		config.NodeName = v.GetString("agent.node")
		config.BindAddr = v.GetString("agent.bind")
		config.AdvertiseAddr = v.GetString("agent.advertise")
		config.StartJoin = v.GetStringSlice("agent.join")
		config.AlertsUrls = v.GetStringSlice("agent.alert_urls")

		markStringRequired(config.NodeName, "node")
		markStringRequired(config.BindAddr, "bind")
		markSliceStringRequired(config.StartJoin, "join")
		markSliceStringRequired(config.AlertsUrls, "alertsUrls")

		return config
	}

	var kind string
	re := regexp.MustCompile("^monitor$|^auditor$|^publisher$")
	for _, arg := range args {
		if re.MatchString(arg) {
			kind = arg
			break
		}
	}

	switch kind {
	case "publisher":
		cmd.AddCommand(newAgentPublisherCommand(cmdCtx, *config, agentPreRun))

	case "auditor":
		cmd.AddCommand(newAgentAuditorCommand(cmdCtx, *config, agentPreRun))

	case "monitor":
		cmd.AddCommand(newAgentMonitorCommand(cmdCtx, *config, agentPreRun))
	}

	return cmd

}
