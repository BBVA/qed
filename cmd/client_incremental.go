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

func newIncrementalCommand(ctx *clientContext) *cobra.Command {

	var start, end uint64
	var verify bool
	var startDigest, endDigest string

	cmd := &cobra.Command{
		Use:   "incremental",
		Short: "Query for incremental",
		Long: `Query for an incremental proof to the authenticated data structure.
			It also verifies the proofs provided by the server if flag enabled.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if verify {
				if startDigest == "" {
					log.Errorf("Error: trying to verify proof without start digest")
				}
				if endDigest == "" {
					log.Errorf("Error: trying to verify proof without end digest")
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Infof("Querying incremental between versions [ %d ] and [ %d ]\n", start, end)

			proof, err := ctx.client.Incremental(start, end)
			if err != nil {
				return err
			}

			log.Infof("Received proof: %+v\n", proof)

			if verify {
				sdBytes, _ := hex.DecodeString(startDigest)
				edBytes, _ := hex.DecodeString(endDigest)
				startSnapshot := &protocol.Snapshot{sdBytes, nil, start, nil}
				endSnapshot := &protocol.Snapshot{edBytes, nil, end, nil}

				log.Infof("Verifying with snapshots: \n\tStartDigest: %s\n\tEndDigest: %s\n",
					startDigest, endDigest)
				if ctx.client.VerifyIncremental(proof, startSnapshot, endSnapshot, hashing.NewSha256Hasher()) {
					log.Info("Verify: OK")
				} else {
					log.Info("Verify: KO")
				}
			}

			return nil
		},
	}

	cmd.Flags().Uint64Var(&start, "start", 0, "Start version to query")
	cmd.Flags().Uint64Var(&end, "end", 0, "End version to query")
	cmd.Flags().BoolVar(&verify, "verify", false, "Do verify received proof")
	cmd.Flags().StringVar(&startDigest, "startDigest", "", "Start digest of the history tree")
	cmd.Flags().StringVar(&endDigest, "endDigest", "", "End digest of the history tree")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}
