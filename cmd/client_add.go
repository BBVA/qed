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
	"os"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/log2"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
)

var clientAddCmd *cobra.Command = &cobra.Command{
	Use:   "add",
	Short: "Add a QED event to the QED log",
	RunE:  runClientAdd,
}

var clientAddCtx context.Context

type addParams struct {
	Event string `desc:"QED event to insert to QED"`
}

func init() {

	clientAddCtx = configClientAdd()
	clientCmd.AddCommand(clientAddCmd)
}

func configClientAdd() context.Context {

	conf := &addParams{}

	err := gpflag.ParseTo(conf, clientAddCmd.PersistentFlags())
	if err != nil {
		fmt.Printf("Cannot parse command flags: %v\n", err)
		os.Exit(1)
	}
	return context.WithValue(Ctx, k("client.add.params"), conf)
}

func runClientAdd(cmd *cobra.Command, args []string) error {

	params := clientAddCtx.Value(k("client.add.params")).(*addParams)

	if params.Event == "" {
		return fmt.Errorf("Event must not be empty!")
	}

	config := clientCtx.Value(k("client.config")).(*client.Config)

	// create main logger
	logOpts := &log2.LoggerOptions{
		Name:            "qed.client",
		IncludeLocation: true,
		Level:           log2.LevelFromString(config.Log),
		Output:          log2.DefaultOutput,
		TimeFormat:      log2.DefaultTimeFormat,
	}
	logger := log2.New(logOpts)

	client, err := client.NewHTTPClientFromConfigWithLogger(config, logger)
	if err != nil {
		return err
	}

	snapshot, err := client.Add(params.Event)
	if err != nil {
		return err
	}

	fmt.Printf("\nReceived snapshot with values:\n\n")
	fmt.Printf(" EventDigest: %x\n", snapshot.EventDigest)
	fmt.Printf(" HistoryDigest: %x\n", snapshot.HistoryDigest)
	fmt.Printf(" HyperDigest: %x\n", snapshot.HyperDigest)
	fmt.Printf(" Version: %d\n\n", snapshot.Version)

	return nil
}
