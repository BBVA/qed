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

var clientBackupCmd *cobra.Command = &cobra.Command{
	Use:   "backup",
	Short: "Make a backup of the QED Log",
	RunE:  runClientBackup,
}

var clientBackupCtx context.Context

type backupParams struct {
	Path string `desc:"Path to save the backup"`
}

func init() {

	clientBackupCtx = configClientBackup()
	clientCmd.AddCommand(clientBackupCmd)
}

func configClientBackup() context.Context {

	conf := &backupParams{}

	err := gpflag.ParseTo(conf, clientBackupCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("client.backup.params"), conf)
}

func runClientBackup(cmd *cobra.Command, args []string) error {

	params := clientBackupCtx.Value(k("client.backup.params")).(*backupParams)

	if params.Path == "" {
		return fmt.Errorf("Path must not be empty!")
	}

	config := clientCtx.Value(k("client.config")).(*client.Config)
	log.SetLogger("client", config.Log)

	client, err := client.NewHTTPClientFromConfig(config)
	if err != nil {
		return err
	}

	err = client.Backup(params.Path)
	if err != nil {
		return err
	}

	fmt.Printf("\nBackup made at %s.\n", params.Path)

	return nil
}
