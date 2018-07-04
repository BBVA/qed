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
	"github.com/spf13/cobra"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/server"
	"github.com/bbva/qed/sign"
)

func newStartCommand() *cobra.Command {
	var (
		endpoint, dbPath, storageName, privateKeyPath string
		cacheSize                                     uint64
		profiling, tampering                          bool
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server for the verifiable log QED",
		Long:  ``,
		// Args:  cobra.NoArgs(),

		Run: func(cmd *cobra.Command, args []string) {

			signer, err := sign.NewSignerFromFile(privateKeyPath)
			if err != nil {
				log.Error(err)
			}

			srv := server.NewServer(
				endpoint,
				dbPath,
				apiKey,
				cacheSize,
				storageName,
				profiling,
				tampering,
				signer,
			)

			err = srv.Run()
			if err != nil {
				log.Errorf("Can't start QED server: %v", err)
			}

		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", "0.0.0.0:8080", "Endpoint for REST requests on (host:port)")
	cmd.Flags().StringVarP(&dbPath, "path", "p", "/var/tmp/qed.db", "Set default storage path.")
	cmd.Flags().Uint64VarP(&cacheSize, "cache", "c", 1<<25, "Initialize and reserve custom cache size.")
	cmd.Flags().StringVarP(&storageName, "storage", "s", "badger", "Choose between different storage backends. Eg badge|bolt")
	cmd.Flags().StringVarP(&privateKeyPath, "keypath", "y", "~/.ssh/id_ed25519", "Path to the ed25519 key file")
	cmd.Flags().BoolVarP(&profiling, "profiling", "f", false, "Allow a pprof url (localhost:6060) for profiling purposes")

	// INFO: testing purposes
	cmd.Flags().BoolVar(&tampering, "tampering", false, "Allow tampering api for proof demostrations")
	cmd.Flags().MarkHidden("tampering")

	return cmd
}
