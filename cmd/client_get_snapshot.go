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
	"fmt"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
)

var clientGetCmd *cobra.Command = &cobra.Command{
	Use:   "get",
	Short: "Get a QED snapshot from a public Snapshot store",
	RunE:  runClientGet,
}

var clientGetCtx context.Context

func init() {
	clientGetCtx = configClientGet()
	clientCmd.AddCommand(clientGetCmd)
}

type getParams struct {
	Version          uint64 `desc:"Snapshot version to look for"`
	SnapshotStoreURL string `desc:"Snapshot store URL"`
}

func configClientGet() context.Context {

	conf := &getParams{}

	err := gpflag.ParseTo(conf, clientGetCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("client.get.params"), conf)
}

func runClientGet(cmd *cobra.Command, args []string) error {

	params := clientGetCtx.Value(k("client.get.params")).(*getParams)

	//TO DO: Check if "version" is set in params. Default value = 0, so it works.

	if params.SnapshotStoreURL == "" {
		fmt.Println("\nUsing default snapshot store: http://127.0.0.1:8888")
	}

	config := clientCtx.Value(k("client.config")).(*client.Config)
	log.SetLogger("client", config.Log)

	client, err := client.NewHTTPClientFromConfig(config)
	if err != nil {
		return err
	}

	snapshot, err := client.GetSnapshot(params.Version)
	if err != nil {
		return err
	}

	fmt.Printf("\nRetreived snapshot with values:\n\n")
	fmt.Printf(" EventDigest: %x\n", snapshot.EventDigest)
	fmt.Printf(" HyperDigest: %x\n", snapshot.HyperDigest)
	fmt.Printf(" HistoryDigest: %x\n", snapshot.HistoryDigest)
	fmt.Printf(" Version: %d\n\n", snapshot.Version)

	return nil
}
