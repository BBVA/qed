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

var clientRestoreCmd *cobra.Command = &cobra.Command{
	Use:   "restore",
	Short: "Restore the QED Log from a backup",
	RunE:  runClientRestore,
}

var clientRestoreCtx context.Context

type restoreParams struct {
	Path string `desc:"Path to find the backup"`
}

func init() {

	clientRestoreCtx = configClientRestore()
	clientCmd.AddCommand(clientRestoreCmd)
}

func configClientRestore() context.Context {

	conf := &backupParams{}

	err := gpflag.ParseTo(conf, clientRestoreCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("client.restore.params"), conf)
}

func runClientRestore(cmd *cobra.Command, args []string) error {

	params := clientRestoreCtx.Value(k("client.restore.params")).(*restoreParams)

	if params.Path == "" {
		return fmt.Errorf("Path must not be empty!")
	}

	config := clientCtx.Value(k("client.config")).(*client.Config)
	log.SetLogger("client", config.Log)

	client, err := client.NewHTTPClientFromConfig(config)
	if err != nil {
		return err
	}

	err = client.Restore(params.Path)
	if err != nil {
		return err
	}

	fmt.Printf("\nRestore from backup at %s done.\n", params.Path)

	return nil
}
