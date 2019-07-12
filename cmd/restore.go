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
	"errors"
	"fmt"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/rocksdb"
)

type RestoreConfig struct {
	// Path where backups are placed.
	BackupDir string `desc:"Path where backups are placed"`

	// Backup to restore.
	BackupID uint32 `desc:"Backup to restore."`

	// Path to restore a backup.
	RestorePath string `desc:"Path to restore a backup"`

	// Log level
	Log string `desc:"Set log level to info, error or debug"`
}

func defaultRestoreConfig() *RestoreConfig {
	return &RestoreConfig{
		BackupDir:   "",
		BackupID:    0,
		RestorePath: "",
	}
}

var restoreCmd *cobra.Command = &cobra.Command{
	Use:              "restore",
	Short:            "Restore a QED log backup",
	TraverseChildren: true,
	RunE:             runRestore,
}

var restoreCtx context.Context

func init() {
	restoreCtx = configRestore()
	Root.AddCommand(restoreCmd)
}

func configRestore() context.Context {

	conf := defaultRestoreConfig()

	err := gpflag.ParseTo(conf, restoreCmd.PersistentFlags())
	if err != nil {
		log.Fatalf("err: %v", err)
	}
	return context.WithValue(Ctx, k("restore.config"), conf)
}

func runRestore(cmd *cobra.Command, args []string) error {

	params := restoreCtx.Value(k("restore.config")).(*RestoreConfig)

	if params.BackupDir == "" {
		return errors.New("Backup directory is empty.")
	}
	if params.RestorePath == "" {
		return errors.New("Restore directory is empty.")
	}

	bo := rocksdb.NewDefaultOptions()
	be, err := rocksdb.OpenBackupEngine(bo, params.BackupDir)
	if err != nil {
		return err
	}

	ro := rocksdb.NewRestoreOptions()

	if params.BackupID == 0 {
		err = be.RestoreDBFromLatestBackup(params.RestorePath, params.RestorePath, ro)
		if err != nil {
			return err
		}
		fmt.Println("Restore from latest backup completed!")
	} else {
		err = be.RestoreDBFromBackup(params.BackupID, params.RestorePath, params.RestorePath, ro)
		if err != nil {
			return err
		}
		fmt.Printf("Restore from backup %d completed!\n", params.BackupID)
	}
	return nil
}
