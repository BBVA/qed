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
	"os"
	"os/signal"
	"syscall"

	"github.com/bbva/qed/gossip"
	"github.com/bbva/qed/gossip/auditor"
	"github.com/bbva/qed/gossip/monitor"
	"github.com/bbva/qed/log"
	"github.com/spf13/cobra"
)

func newAgentStartCommand() *cobra.Command {
	var (
		nodeName, bindAddr, advertiseAddr, role string
		startJoin                               []string
	)

	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Start a gossip agent for the verifiable log QED",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {

			config := gossip.DefaultConfig()
			config.NodeName = nodeName
			config.BindAddr = bindAddr
			config.AdvertiseAddr = advertiseAddr
			config.Role = gossip.NewNodeType(role)
			config.EnableCompression = true

			var agent *gossip.Agent
			var err error
			switch config.Role {
			case gossip.PublisherType:
			case gossip.MonitorType:
				conf := monitor.DefaultConfig()
				agent, err = gossip.Create(config, monitor.NewMonitorHandlerBuilder(conf))
			case gossip.AuditorType:
				conf := auditor.DefaultConfig()
				agent, err = gossip.Create(config, auditor.NewAuditorHandlerBuilder(conf))
			default:
				log.Fatalf("Failed to start the QED agent: unknown role")
			}
			if err != nil {
				log.Fatalf("Failed to start the QED agent: %v", err)
			}

			contacted, err := agent.Join(startJoin)
			if err != nil {
				log.Fatalf("Failed to join the cluster: %v", err)
			}
			log.Debugf("Number of nodes contacted: %d", contacted)

			defer agent.Shutdown()
			awaitTermSignal(agent.Leave)

		},
	}

	cmd.Flags().StringVarP(&nodeName, "node", "", "", "Unique name for node. If not set, fallback to hostname")
	cmd.Flags().StringVarP(&bindAddr, "bind", "", "", "Bind address for TCP/UDP gossip on (host:port)")
	cmd.Flags().StringVarP(&advertiseAddr, "advertise", "", "", "Address to advertise to cluster")
	cmd.Flags().StringVarP(&role, "role", "", "", "Role name")
	cmd.Flags().StringSliceVarP(&startJoin, "join", "", []string{}, "Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")

	return cmd

}

func awaitTermSignal(closeFn func() error) {

	signals := make(chan os.Signal, 1)
	// sigint: Ctrl-C, sigterm: kill command
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// block main and wait for a signal
	sig := <-signals
	log.Infof("Signal received: %v", sig)

	closeFn()

}
