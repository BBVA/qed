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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
)

func TestBackup(t *testing.T) {
	var err error
	log.SetLogger("TestBackup", log.SILENT)

	// New raft node
	raftNode, clean, err := newSeed(0)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, raftNode.Close(true))
		clean()
	}()

	// Insert an event
	h := hashing.NewSha256Hasher()
	event := h.Do([]byte("All's right with the world"))
	cmd := newCommand(addEventCommandType)
	_ = cmd.encode(event)
	rl := newLog(1, 1, cmd.data)

	raftNode.Apply(rl)

	// Create backup
	backupsList := raftNode.ListBackups()
	require.True(t, len(backupsList) == 0, "Backup list should be empty")
	err = raftNode.CreateBackup()
	require.NoError(t, err)
	backupsList = raftNode.ListBackups()
	require.True(t, len(backupsList) == 1, "Backup list should contain 1 backup")
}

func TestDeleteBackup(t *testing.T) {
	var err error
	log.SetLogger("TestDeleteBackup", log.SILENT)

	// New raft node
	raftNode, clean, err := newSeed(1)
	defer func() {
		require.NoError(t, raftNode.Close(true))
		clean()
	}()

	// Insert an event
	h := hashing.NewSha256Hasher()
	event := h.Do([]byte("All's right with the world"))
	cmd := newCommand(addEventCommandType)
	_ = cmd.encode(event)
	rl := newLog(1, 1, cmd.data)

	raftNode.Apply(rl)

	// Create backup and delete later.
	backupsList := raftNode.ListBackups()
	require.True(t, len(backupsList) == 0, "Backup list must be empty")
	err = raftNode.CreateBackup()
	require.NoError(t, err)
	backupsList = raftNode.ListBackups()
	require.True(t, len(backupsList) == 1, "Backup list must contain 1 backup")
	err = raftNode.DeleteBackup(1)
	require.NoError(t, err)
	backupsList = raftNode.ListBackups()
	require.True(t, len(backupsList) == 0, "Backup list must be empty again")
	err = raftNode.DeleteBackup(12345)
	require.Error(t, err, "Deleting an unknown backup must return an error")
}
