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
	"fmt"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/protocol"

	"github.com/spf13/cobra"
)

func newIncrementalCommand(ctx *clientContext, clientPreRun func(*cobra.Command, []string)) *cobra.Command {

	var start, end uint64
	var verify bool

	cmd := &cobra.Command{
		Use:   "incremental",
		Short: "Query for incremental",
		Long: `Query for an incremental proof to the authenticated data structure.
			It also verifies the proofs provided by the server if flag enabled.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// WARN: PersitentPreRun can't be nested and we're using it in
			// cmd/root so inbetween preRuns must be curried.
			clientPreRun(cmd, args)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {

			fmt.Printf("\nQuerying incremental between versions [ %d ] and [ %d ]\n", start, end)
			// SilenceUsage is set to true -> https://github.com/spf13/cobra/issues/340
			cmd.SilenceUsage = true
			proof, err := ctx.client.Incremental(start, end)
			if err != nil {
				return err
			}

			fmt.Printf("\nReceived incremental proof: \n\n")
			fmt.Printf(" Start version: %d\n", proof.Start)
			fmt.Printf(" End version: %d\n", proof.End)
			fmt.Printf(" Incremental audit path: <TRUNCATED>\n\n")

			if verify {

				var startDigest, endDigest string
				for {
					startDigest = readLine(fmt.Sprintf("Please, provide the starting historyDigest for version [ %d ]: ", start))
					if startDigest != "" {
						break
					}
				}
				for {
					endDigest = readLine(fmt.Sprintf("Please, provide the ending historyDigest for version [ %d ] : ", end))
					if endDigest != "" {
						break
					}
				}

				sdBytes, _ := hex.DecodeString(startDigest)
				edBytes, _ := hex.DecodeString(endDigest)
				startSnapshot := &protocol.Snapshot{sdBytes, nil, start, nil}
				endSnapshot := &protocol.Snapshot{edBytes, nil, end, nil}

				fmt.Printf("\nVerifying with snapshots: \n")
				fmt.Printf(" HistoryDigest for start version [ %d ]: %s\n", start, startDigest)
				fmt.Printf(" HistoryDigest for end version [ %d ]: %s\n", end, endDigest)

				if ctx.client.VerifyIncremental(proof, startSnapshot, endSnapshot, hashing.NewSha256Hasher()) {
					fmt.Printf("\nVerify: OK\n\n")
				} else {
					fmt.Printf("\nVerify: KO\n\n")
				}
			}

			return nil
		},
	}

	cmd.Flags().Uint64Var(&start, "start", 0, "Start version to query")
	cmd.Flags().Uint64Var(&end, "end", 0, "End version to query")
	cmd.Flags().BoolVar(&verify, "verify", false, "Do verify received proof")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}
