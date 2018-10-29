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

	"github.com/spf13/cobra"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
)

func newStartCommand() *cobra.Command {
	var (
		nodeId, httpAddr, raftAddr, mgmtAddr, joinAddr string
		dbPath, raftPath, privateKeyPath               string
		profiling, tampering, publish                  bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server for the verifiable log QED",
		Long:  ``,
		// Args:  cobra.NoArgs(),

		Run: func(cmd *cobra.Command, args []string) {

			srv, err := server.NewServer(
				nodeId,
				httpAddr,
				raftAddr,
				mgmtAddr,
				joinAddr,
				dbPath,
				raftPath,
				privateKeyPath,
				apiKey,
				profiling,
				tampering,
				publish,
			)

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
	cmd.Flags().StringVarP(&nodeId, "node-id", "", hostname, "Unique name for node. If not set, fallback to hostname")
	cmd.Flags().StringVarP(&httpAddr, "http-addr", "", ":8080", "Endpoint for REST requests on (host:port)")
	cmd.Flags().StringVarP(&raftAddr, "raft-addr", "", ":9000", "Raft bind address (host:port)")
	cmd.Flags().StringVarP(&mgmtAddr, "mgmt-addr", "", ":8090", "Management endpoint bind address (host:port)")
	cmd.Flags().StringVarP(&joinAddr, "join-addr", "", "", "Comma-delimited list of nodes ([host]:port), through which a cluster can be joined")
	cmd.Flags().StringVarP(&dbPath, "dbpath", "p", "/var/tmp/qed/data", "Set default storage path")
	cmd.Flags().StringVarP(&raftPath, "raftpath", "", "/var/tmp/qed/raft", "Set raft storage path")
	cmd.Flags().StringVarP(&privateKeyPath, "keypath", "y", "~/.ssh/id_ed25519", "Path to the ed25519 key file")
	cmd.Flags().BoolVarP(&profiling, "profiling", "f", false, "Allow a pprof url (localhost:6060) for profiling purposes")
	cmd.Flags().BoolVarP(&publish, "publish", "", false, "Enable/Disable publishing snapshots.")

	// INFO: testing purposes
	cmd.Flags().BoolVar(&tampering, "tampering", false, "Allow tampering api for proof demostrations")
	cmd.Flags().MarkHidden("tampering")

	return cmd
}
