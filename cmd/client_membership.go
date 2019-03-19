/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

func newMembershipCommand(ctx *clientContext, clientPreRun func(*cobra.Command, []string)) *cobra.Command {

	hasherF := hashing.NewSha256Hasher
	var version uint64
	var verify bool
	var key, eventDigest string

	cmd := &cobra.Command{
		Use:   "membership",
		Short: "Query for membership",
		Long: `Query for membership of an event to the authenticated data structure.
			It also verifies the proofs provided by the server if flag enabled.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// WARN: PersitentPreRun can't be nested and we're using it in
			// cmd/root so inbetween preRuns must be curried.
			clientPreRun(cmd, args)

			if key == "" && eventDigest == "" {
				log.Errorf("Error: trying to get membership without either key or eventDigest")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var membershipResult *protocol.MembershipResult
			var digest hashing.Digest
			var err error
			// SilenceUsage is set to true -> https://github.com/spf13/cobra/issues/340
			cmd.SilenceUsage = true

			if eventDigest == "" {
				fmt.Printf("\nQuerying key [ %s ] with version [ %d ]\n", key, version)
				digest = hasherF().Do([]byte(key))
			} else {
				fmt.Printf("\nQuerying digest [ %s ] with version [ %d ]\n", eventDigest, version)
				digest, _ = hex.DecodeString(eventDigest)
			}

			membershipResult, err = ctx.client.MembershipDigest(digest, version)
			if err != nil {
				return err
			}
			fmt.Printf("\nReceived membership proof:\n")
			fmt.Printf("\n Exists: %t\n", membershipResult.Exists)
			fmt.Printf(" Hyper audit path: <TRUNCATED>\n")
			fmt.Printf(" History audit path: <TRUNCATED>\n")
			fmt.Printf(" CurrentVersion: %d\n", membershipResult.CurrentVersion)
			fmt.Printf(" QueryVersion: %d\n", membershipResult.QueryVersion)
			fmt.Printf(" ActualVersion: %d\n", membershipResult.ActualVersion)
			fmt.Printf(" KeyDigest: %x\n\n", membershipResult.KeyDigest)

			if verify {

				var hyperDigest, historyDigest string
				for {
					hyperDigest = readLine(fmt.Sprintf("Please, provide the hyperDigest for current version [ %d ]: ", membershipResult.CurrentVersion))
					if hyperDigest != "" {
						break
					}
				}
				if membershipResult.Exists {
					for {
						historyDigest = readLine(fmt.Sprintf("Please, provide the historyDigest for version [ %d ] : ", version))
						if historyDigest != "" {
							break
						}
					}
				}

				hdBytes, _ := hex.DecodeString(hyperDigest)
				htdBytes, _ := hex.DecodeString(historyDigest)
				snapshot := &protocol.Snapshot{
					HistoryDigest: htdBytes,
					HyperDigest:   hdBytes,
					Version:       version,
					EventDigest:   digest}

				fmt.Printf("\nVerifying with Snapshot: \n\n EventDigest:%x\n HyperDigest: %s\n HistoryDigest: %s\n Version: %d\n",
					digest, hyperDigest, historyDigest, version)

				if ctx.client.DigestVerify(membershipResult, snapshot, hasherF) {
					fmt.Printf("\nVerify: OK\n\n")
				} else {
					fmt.Printf("\nVerify: KO\n\n")
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Key to query")
	cmd.Flags().Uint64Var(&version, "version", 0, "Version to query")
	cmd.Flags().BoolVar(&verify, "verify", false, "Do verify received proof")
	cmd.Flags().StringVar(&eventDigest, "eventDigest", "", "Digest of the event")

	cmd.MarkFlagRequired("version")

	return cmd
}

func readLine(query string) string {
	fmt.Print(query)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	// convert CRLF to LF
	return strings.Replace(text, "\n", "", -1)
}
