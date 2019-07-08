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

	"github.com/bbva/qed/log"
)

type BackupConfig struct {
	// Endpoint [host:port] to ask for QED management APIs.
	Endpoint string `desc:"QED Log service management endpoint http://ip:port"`

	// Log level
	Log string `desc:"Set log level to info, error or debug"`

	// ApiKey to query the server endpoint.
	APIKey string `desc:"Set API Key to talk to QED Log service"`
}

func defaultConfig() *BackupConfig {
	return &BackupConfig{
		Endpoint: "http://127.0.0.1:8700",
		APIKey:   "my-key",
	}
}

var backupCmd *cobra.Command = &cobra.Command{
	Use:               "backup",
	Short:             "Manages QED log backups",
	TraverseChildren:  true,
	PersistentPreRunE: runBackup,
}

var backupCtx context.Context

func init() {
	backupCtx = configBackup()
	Root.AddCommand(backupCmd)
}

func configBackup() context.Context {

	conf := defaultConfig()

	err := gpflag.ParseTo(conf, backupCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("backup.config"), conf)
}

func runBackup(cmd *cobra.Command, args []string) error {
	var err error

	endpoint, _ := cmd.Flags().GetString("endpoint")
	err = urlParse(endpoint)
	if err != nil {
		return err
	}

	return nil
}
