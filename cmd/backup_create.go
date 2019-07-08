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

	"github.com/spf13/cobra"

	"github.com/bbva/qed/log"
)

var backupCreateCmd *cobra.Command = &cobra.Command{
	Use:   "create",
	Short: "Create a backup of the QED Log",
	RunE:  runBackupCreate,
}

var backupCreateCtx context.Context

func init() {
	backupCmd.AddCommand(backupCreateCmd)
}

func runBackupCreate(cmd *cobra.Command, args []string) error {

	config := backupCtx.Value(k("backup.config")).(*BackupConfig)
	log.SetLogger("backup", config.Log)

	_, err := createBackup(config)
	if err != nil {
		return err
	}

	fmt.Println("Backup created!")

	return nil
}

func createBackup(config *BackupConfig) ([]byte, error) {

	// Build request
	req, err := http.NewRequest("POST", config.Endpoint+"/backup", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", config.APIKey)

	// Get response
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Infof("Request error: %v\n", err)
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
