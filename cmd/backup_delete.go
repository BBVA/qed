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
	"io/ioutil"
	"net/http"
	"os"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
)

var backupDeleteCmd *cobra.Command = &cobra.Command{
	Use:   "delete",
	Short: "Delete a QED Log backup",
	RunE:  runBackupDelete,
}

var backupDeleteCtx context.Context

type deleteParams struct {
	BackupID uint32 `desc:"QED backup to delete"`
}

func init() {
	backupDeleteCtx = configBackupDelete()
	backupCmd.AddCommand(backupDeleteCmd)
}

func configBackupDelete() context.Context {
	conf := &deleteParams{}

	err := gpflag.ParseTo(conf, backupDeleteCmd.PersistentFlags())
	if err != nil {
		fmt.Printf("Cannot parse command flags: %v\n", err)
		os.Exit(1)
	}
	return context.WithValue(Ctx, k("backup.delete.params"), conf)
}

func runBackupDelete(cmd *cobra.Command, args []string) error {
	params := backupDeleteCtx.Value(k("backup.delete.params")).(*deleteParams)

	config := backupCtx.Value(k("backup.config")).(*BackupConfig)

	_, err := deleteBackup(config, params.BackupID)
	if err != nil {
		return err
	}

	fmt.Println("Backup deleted!")
	return nil
}

func deleteBackup(config *BackupConfig, backupID uint32) ([]byte, error) {

	// Build request
	req, err := http.NewRequest("DELETE", config.Endpoint+"/backup?backupID="+fmt.Sprintf("%d", backupID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", config.APIKey)

	// Get response
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request error: %v\n", err)
		return nil, err
	}

	var bodyBytes []byte
	if resp.Body != nil {
		defer resp.Body.Close()
		bodyBytes, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return nil, fmt.Errorf("Invalid request %v", string(bodyBytes))
	}

	return bodyBytes, nil
}
