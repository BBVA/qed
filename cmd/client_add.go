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

	"github.com/spf13/cobra"
)

func newAddCommand(ctx *clientContext, clientPreRun func(*cobra.Command, []string)) *cobra.Command {

	var key string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an event",
		Long:  `Add an event to the authenticated data structure`,
		PreRun: func(cmd *cobra.Command, args []string) {
			// WARN: PersitentPreRun can't be nested and we're using it in
			// cmd/root so inbetween preRuns must be curried.
			clientPreRun(cmd, args)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("\nAdding key [ %s ]\n", key)
			// SilenceUsage is set to true -> https://github.com/spf13/cobra/issues/340
			cmd.SilenceUsage = true
			snapshot, err := ctx.client.Add(key)
			if err != nil {
				return err
			}

			fmt.Printf("\nReceived snapshot with values:\n\n")
			fmt.Printf(" EventDigest: %x\n", snapshot.EventDigest)
			fmt.Printf(" HyperDigest: %x\n", snapshot.HyperDigest)
			fmt.Printf(" HistoryDigest: %x\n", snapshot.HistoryDigest)
			fmt.Printf(" Version: %d\n\n", snapshot.Version)

			return nil
		},
	}

	cmd.Flags().StringVar(&key, "key", "", "Key to add")
	cmd.MarkFlagRequired("key")

	return cmd
}
