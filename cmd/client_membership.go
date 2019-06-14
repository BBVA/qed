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
	"context"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/client"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
)

var clientMembershipCmd *cobra.Command = &cobra.Command{
	Use:   "membership",
	Short: "Query for membership",
	Long: `Query for membership of an event to the authenticated data structure.
It also verifies the proofs provided by the server if flag enabled.`,
	RunE: runClientMembership,
}

var clientMembershipCtx context.Context

func init() {
	clientMembershipCtx = configClientMembership()
	clientCmd.AddCommand(clientMembershipCmd)
}

type membershipParams struct {
	Version       *uint64 `desc:"Version for the membership proof"`
	Event         string  `desc:"QED event to build the proof"`
	EventDigest   string  `desc:"QED event digest to build the proof"`
	HistoryDigest string  `desc:"QED history digest is used to verify the proof"`
	HyperDigest   string  `desc:"QED hyper digest is used to verify the proof"`
	Verify        bool    `desc:"Set to enable proof verification process"`
	AutoVerify    bool    `desc:"Set to enable proof automatic verification process"`
}

func configClientMembership() context.Context {

	conf := &membershipParams{}
	err := gpflag.ParseTo(conf, clientMembershipCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("client.membership.params"), conf)
}

func checkVersionSet(cmd *cobra.Command) bool {
	var found bool
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "version" {
			found = f.Changed
		}
	})
	return found
}

func runClientMembership(cmd *cobra.Command, args []string) error {

	hasherF := hashing.NewSha256Hasher

	var proof *balloon.MembershipProof
	var digest hashing.Digest
	var err error
	var msg string

	params := clientMembershipCtx.Value(k("client.membership.params")).(*membershipParams)

	// SilenceUsage is set to true -> https://github.com/spf13/cobra/issues/340
	cmd.SilenceUsage = true

	if params.EventDigest == "" {
		msg += fmt.Sprintf("Querying key [ %s ]", params.Event)
		digest = hasherF().Do([]byte(params.Event))
	} else {
		msg += fmt.Sprintf("Querying digest [ %s ]", params.EventDigest)
		digest, _ = hex.DecodeString(params.EventDigest)
	}

	if !checkVersionSet(cmd) {
		params.Version = nil
		msg += " with latest version"
	} else {
		msg += fmt.Sprintf(" with version [ %d ]", *params.Version)

	}

	fmt.Printf("\n%s\n", msg)

	config := clientCtx.Value(k("client.config")).(*client.Config)

	client, err := client.NewHTTPClientFromConfig(config)
	if err != nil {
		return err
	}

	proof, err = client.MembershipDigest(digest, params.Version)
	if err != nil {
		return err
	}
	fmt.Printf("\nReceived membership proof:\n\n")
	fmt.Printf(" Exists: %t\n", proof.Exists)
	fmt.Printf(" Hyper audit path: <TRUNCATED>\n")
	fmt.Printf(" History audit path: <TRUNCATED>\n")
	fmt.Printf(" CurrentVersion: %d\n", proof.CurrentVersion)
	fmt.Printf(" QueryVersion: %d\n", proof.QueryVersion)
	fmt.Printf(" ActualVersion: %d\n", proof.ActualVersion)
	fmt.Printf(" KeyDigest: %x\n\n", proof.KeyDigest)

	if params.AutoVerify || params.Verify {
		var ok bool
		var err error

		if params.AutoVerify {
			fmt.Printf("\nAuto-Verifying event with: \n\n EventDigest: %x\n Version: %d\n", digest, proof.QueryVersion)
			ok, err = client.MembershipAutoVerify(digest, params.Version)
		} else {

			hyperDigest := params.HyperDigest
			historyDigest := params.HistoryDigest
			for hyperDigest == "" {
				hyperDigest = readLine(fmt.Sprintf("Please, provide the hyperDigest for current version [ %d ]: ", proof.CurrentVersion))
			}
			if proof.Exists {
				for historyDigest == "" {
					historyDigest = readLine(fmt.Sprintf("Please, provide the historyDigest for version [ %d ] : ", proof.QueryVersion))
				}
			}
			hdBytes, _ := hex.DecodeString(hyperDigest)
			htdBytes, _ := hex.DecodeString(historyDigest)

			snapshot := &balloon.Snapshot{
				HistoryDigest: htdBytes,
				HyperDigest:   hdBytes,
				Version:       uint64(0),
				EventDigest:   digest,
			}

			fmt.Printf("\nVerifying event with: \n\n EventDigest: %x\n HyperDigest: %x\n HistoryDigest: %x\n Version: %d\n", digest, hdBytes, htdBytes, proof.QueryVersion)
			ok, err = client.MembershipVerify(digest, proof, snapshot)
		}

		if ok {
			fmt.Printf("\nVerify: OK\n\n")
		} else {
			fmt.Printf("\nVerify: KO\n\n")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func readLine(query string) string {
	fmt.Print(query)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	// convert CRLF to LF
	return strings.Replace(text, "\n", "", -1)
}
