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

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
)

var clientCmd *cobra.Command = &cobra.Command{
	Use:               "client",
	Short:             "Provdes access to the QED log client",
	TraverseChildren:  true,
	PersistentPreRunE: runClient,
}

var clientCtx context.Context

func init() {
	clientCtx = configClient()
	Root.AddCommand(clientCmd)
}

func configClient() context.Context {

	conf := client.DefaultConfig()

	err := gpflag.ParseTo(conf, clientCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("client.config"), conf)
}

func runClient(cmd *cobra.Command, args []string) error {
	var err error

	endpoints, _ := cmd.Flags().GetStringSlice("endpoints")
	err = urlParse(endpoints...)
	if err != nil {
		return err
	}

	snapshotStoreURL, _ := cmd.Flags().GetString("snapshot-store-url")
	err = urlParse(snapshotStoreURL)
	if err != nil {
		return err
	}

	return nil
}
