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
	"encoding/hex"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/protocol"

	"github.com/spf13/cobra"

	"github.com/bbva/qed/log"
)

func newMembershipCommand(ctx *clientContext) *cobra.Command {

	var version uint64
	var verify bool
	var key, hyperDigest, historyDigest string

	cmd := &cobra.Command{
		Use:   "membership",
		Short: "Query for membership",
		Long: `Query for membership of an event to the authenticated data structure.
			It also verifies the proofs provided by the server if flag enabled.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if verify {
				if hyperDigest == "" {
					log.Errorf("Error: trying to verify proof without hyper digest")
				}
				if historyDigest == "" {
					log.Errorf("Error: trying to verify proof without history digest")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Querying key [ %s ] with version [ %d ]\n", key, version)

			event := []byte(key)
			proof, err := ctx.client.Membership(event, version)
			if err != nil {
				return err
			}

			log.Infof("Received proof: %+v\n", proof)

			if verify {
				hdBytes, _ := hex.DecodeString(hyperDigest)
				htdBytes, _ := hex.DecodeString(historyDigest)
				snapshot := &protocol.Snapshot{htdBytes, hdBytes, version, event}

				log.Infof("Verifying with Snapshot: \n\tEventDigest:%s\n\tHyperDigest: %s\n\tHistoryDigest: %s\n\tVersion: %d\n",
					event, hyperDigest, historyDigest, version)
				if ctx.client.Verify(proof, snapshot, hashing.NewSha256Hasher) {
					log.Info("Verify: OK")
				} else {
					log.Info("Verify: KO")
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Key to query")
	cmd.Flags().Uint64Var(&version, "version", 0, "Version to query")
	cmd.Flags().BoolVar(&verify, "verify", false, "Do verify received proof")
	cmd.Flags().StringVar(&hyperDigest, "hyperDigest", "", "Digest of the hyper tree")
	cmd.Flags().StringVar(&historyDigest, "historyDigest", "", "Digest of the history tree")

	cmd.MarkFlagRequired("key")
	cmd.MarkFlagRequired("version")

	return cmd
}
