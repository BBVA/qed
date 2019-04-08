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
	"fmt"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
	"github.com/spf13/cobra"
)

var clientAddCmd *cobra.Command = &cobra.Command{
	Use:   "add",
	Short: "Add a QED event to the QED log",
	RunE:  runClientAdd,
}

var clientAddEvent string

func init() {

	clientAddCmd.Flags().StringVar(&clientAddEvent, "event", "", "Event to append to QED")
	clientAddCmd.MarkFlagRequired("event")

	clientCmd.AddCommand(clientAddCmd)
}

func runClientAdd(cmd *cobra.Command, args []string) error {
	// SilenceUsage is set to true -> https://github.com/spf13/cobra/issues/340
	if clientAddEvent == "" {
		return fmt.Errorf("Event must not be empty!")
	}

	cmd.SilenceUsage = true

	config := clientCtx.Value(k("client.config")).(*client.Config)
	log.SetLogger("client", config.Log)

	client, err := client.NewHTTPClientFromConfig(config)
	if err != nil {
		return err
	}

	snapshot, err := client.Add(clientAddEvent)
	if err != nil {
		return err
	}

	fmt.Printf("\nReceived snapshot with values:\n\n")
	fmt.Printf(" EventDigest: %x\n", snapshot.EventDigest)
	fmt.Printf(" HyperDigest: %x\n", snapshot.HyperDigest)
	fmt.Printf(" HistoryDigest: %x\n", snapshot.HistoryDigest)
	fmt.Printf(" Version: %d\n\n", snapshot.Version)

	return nil
}

