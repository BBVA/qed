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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/bbva/qed/protocol"
)

var backupListCmd *cobra.Command = &cobra.Command{
	Use:   "list",
	Short: "List QED Log backups",
	RunE:  runBackupList,
}

var backupListCtx context.Context

func init() {
	backupCmd.AddCommand(backupListCmd)
}

func runBackupList(cmd *cobra.Command, args []string) error {

	config := backupCtx.Value(k("backup.config")).(*BackupConfig)

	listBackupsInfo, err := listBackups(config)
	if err != nil {
		return err
	}

	fmt.Println("Backup list:")
	for _, b := range listBackupsInfo {
		sizeInGB := b.Size / (1024 * 1024 * 1024)
		fmt.Printf("Id: %d\tTimestamp: %s\tVersion: %s\tSize(GB): %d\tNum.Files: %d\t \n", b.ID, formatTimestamp(b.Timestamp), b.Metadata, sizeInGB, b.NumFiles)
	}
	return nil
}

func listBackups(config *BackupConfig) ([]protocol.BackupInfo, error) {

	// Build request
	req, err := http.NewRequest("GET", config.Endpoint+"/backups", nil)
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

	var backupInfo []protocol.BackupInfo
	err = json.Unmarshal(bodyBytes, &backupInfo)
	if err != nil {
		return nil, err
	}

	return backupInfo, nil
}

func formatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0)
	// return t.String()
	return fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}
