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
	"fmt"
	"os"
	"os/user"

	"github.com/spf13/cobra"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
)

func newStartCommand() *cobra.Command {
	const defaultKeyPath = "~/.ssh/id_ed25519"

	conf := server.DefaultConfig()

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server for the verifiable log QED",
		Long:  ``,
		// Args:  cobra.NoArgs(),

		Run: func(cmd *cobra.Command, args []string) {

			if conf.PrivateKeyPath == defaultKeyPath {
				usr, _ := user.Current()
				conf.PrivateKeyPath = fmt.Sprintf("%s/.ssh/id_ed25519", usr.HomeDir)
			}

			srv, err := server.NewServer(conf)

			if err != nil {
				log.Fatalf("Can't start QED server: %v", err)
			}

			err = srv.Start()
			if err != nil {
				log.Fatalf("Can't start QED server: %v", err)
			}

		},
	}

	hostname, _ := os.Hostname()
	cmd.Flags().StringVar(&conf.NodeID, "node-id", hostname, "Unique name for node. If not set, fallback to hostname")
	cmd.Flags().StringVar(&conf.HttpAddr, "http-addr", ":8080", "Endpoint for REST requests on (host:port)")
	cmd.Flags().StringVar(&conf.RaftAddr, "raft-addr", ":9000", "Raft bind address (host:port)")
	cmd.Flags().StringVar(&conf.MgmtAddr, "mgmt-addr", ":8090", "Management endpoint bind address (host:port)")
	cmd.Flags().StringSliceVar(&conf.RaftJoinAddr, "join-addr", []string{}, "Raft: Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")
	cmd.Flags().StringVar(&conf.GossipAddr, "gossip-addr", ":9100", "Gossip: management endpoint bind address (host:port)")
	cmd.Flags().StringSliceVar(&conf.GossipJoinAddr, "gossip-join-addr", []string{}, "Gossip: Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")
	cmd.Flags().StringVarP(&conf.DBPath, "dbpath", "p", "/var/tmp/qed/data", "Set default storage path")
	cmd.Flags().StringVar(&conf.RaftPath, "raftpath", "/var/tmp/qed/raft", "Set raft storage path")
	cmd.Flags().StringVarP(&conf.PrivateKeyPath, "keypath", "y", defaultKeyPath, "Path to the ed25519 key file")
	cmd.Flags().BoolVarP(&conf.EnableProfiling, "profiling", "f", false, "Allow a pprof url (localhost:6060) for profiling purposes")

	// INFO: testing purposes
	cmd.Flags().BoolVar(&conf.EnableTampering, "tampering", false, "Allow tampering api for proof demostrations")
	cmd.Flags().MarkHidden("tampering")

	return cmd
}
