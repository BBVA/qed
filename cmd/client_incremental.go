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
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log2"
	"github.com/octago/sflags/gen/gpflag"

	"github.com/spf13/cobra"
)

var clientIncrementalCmd *cobra.Command = &cobra.Command{
	Use:   "incremental",
	Short: "Query for incremental proof",
	Long: `Query for an incremental proof to the authenticated data structure.
It also verifies the proofs provided by the server if flag enabled.`,
	RunE: runClientIncremental,
}

var clientIncrementalCtx context.Context

func init() {
	clientIncrementalCtx = configClientIncremental()
	clientCmd.AddCommand(clientIncrementalCmd)
}

type incrementalParams struct {
	Start      uint64 `desc:"Starting version for the incremental proof"`
	End        uint64 `desc:"Ending version for the incremental proof"`
	Verify     bool   `desc:"Set to enable proof verification process"`
	AutoVerify bool   `desc:"Set to enable proof automatic verification process"`
}

func configClientIncremental() context.Context {

	conf := &incrementalParams{}

	err := gpflag.ParseTo(conf, clientIncrementalCmd.PersistentFlags())
	if err != nil {
		fmt.Printf("Cannot parse command flags: %v\n", err)
		os.Exit(1)
	}
	return context.WithValue(Ctx, k("client.incremental.params"), conf)
}

func runClientIncremental(cmd *cobra.Command, args []string) error {

	// SilenceUsage is set to true -> https://github.com/spf13/cobra/issues/340
	cmd.SilenceUsage = true
	params := clientIncrementalCtx.Value(k("client.incremental.params")).(*incrementalParams)
	fmt.Printf("\nQuerying incremental between versions [ %d ] and [ %d ]\n", params.Start, params.End)

	clientConfig := clientCtx.Value(k("client.config")).(*client.Config)

	// create main logger
	logOpts := &log2.LoggerOptions{
		Name:            "qed.client",
		IncludeLocation: true,
		Level:           log2.LevelFromString(clientConfig.Log),
		Output:          log2.DefaultOutput,
		TimeFormat:      log2.DefaultTimeFormat,
	}
	logger := log2.New(logOpts)

	client, err := client.NewHTTPClientFromConfigWithLogger(clientConfig, logger)
	if err != nil {
		return err
	}

	proof, err := client.Incremental(params.Start, params.End)
	if err != nil {
		return err
	}

	fmt.Printf("\nReceived incremental proof: \n\n")
	fmt.Printf(" Start version: %d\n", proof.Start)
	fmt.Printf(" End version: %d\n", proof.End)
	fmt.Printf(" Incremental audit path: <TRUNCATED>\n\n")

	if params.AutoVerify || params.Verify {
		var ok bool
		var err error

		if params.AutoVerify {
			fmt.Printf("\nAuto-Verifying event with: \n\n Start: %d\n End: %d\n", params.Start, params.End)
			ok, err = client.IncrementalAutoVerify(params.Start, params.End)
		} else {

			var startDigest, endDigest string
			for {
				startDigest = readLine(fmt.Sprintf("Please, provide the starting historyDigest for version [ %d ]: ", params.Start))
				if startDigest != "" {
					break
				}
			}
			for {
				endDigest = readLine(fmt.Sprintf("Please, provide the ending historyDigest for version [ %d ] : ", params.End))
				if endDigest != "" {
					break
				}
			}

			sdBytes, _ := hex.DecodeString(startDigest)
			edBytes, _ := hex.DecodeString(endDigest)
			startSnapshot := &balloon.Snapshot{
				EventDigest:   nil,
				HistoryDigest: sdBytes,
				HyperDigest:   nil,
				Version:       params.Start,
			}
			endSnapshot := &balloon.Snapshot{
				EventDigest:   nil,
				HistoryDigest: edBytes,
				HyperDigest:   nil,
				Version:       params.End,
			}

			fmt.Printf("\nVerifying with snapshots: \n")
			fmt.Printf(" HistoryDigest for start version [ %d ]: %s\n", params.Start, startDigest)
			fmt.Printf(" HistoryDigest for end version [ %d ]: %s\n", params.End, endDigest)

			ok, err = client.IncrementalVerify(proof, startSnapshot, endSnapshot)
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
