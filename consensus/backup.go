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

package consensus

import (
	"fmt"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

// Backup function calls store's backup function, passing certain metadata.
// Previously, it gets balloon version to build this metadata.
func (n *RaftNode) CreateBackup() error {
	n.Lock()
	defer n.Unlock()

	v := n.balloon.Version()
	metadata := fmt.Sprintf("%d", v-1)
	err := n.db.Backup(metadata)
	if err != nil {
		return err
	}
	log.Debugf("Generating backup until version: %d", v-1)

	return nil
}

// DeleteBackup function is a passthough to store's equivalent funcion.
func (n *RaftNode) DeleteBackup(backupID uint32) error {
	log.Debugf("Deleting backup %d", backupID)
	return n.db.DeleteBackup(backupID)
}

// BackupsInfo function is a passthough to store's equivalent funcion.
func (n *RaftNode) ListBackups() []*storage.BackupInfo {
	log.Debugf("Retrieving backups information")
	return n.db.GetBackupsInfo()
}
