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

package rocksdb

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func insertKeys(db *DB, total, from int) error {

	wo := NewDefaultWriteOptions()
	for i := from; i < from+total; i++ {
		err := db.Put(wo, []byte("key"+string(i)), []byte("value"))
		if err != nil {
			return err
		}
	}
	wo.Destroy()

	return nil
}

// Test from:
// https://github.com/facebook/rocksdb/wiki/How-to-backup-RocksDB%3F#creating-and-verifying-a-backup
func TestBackupAndRestore(t *testing.T) {
	var err error

	backupDir, err := ioutil.TempDir("", "backup")
	require.NoError(t, err)
	err = os.RemoveAll(backupDir)
	require.NoError(t, err)

	// Create new DB and insert keys
	db, _ := newTestDB(t, "original", nil)
	err = insertKeys(db, 10, 0)
	require.NoError(t, err, "Error inserting keys")

	fo := NewDefaultFlushOptions()
	db.Flush(fo)
	fo.Destroy()

	// Create a backup engine
	opts := NewDefaultOptions()
	be, err := OpenBackupEngine(opts, backupDir)
	require.NoError(t, err)
	require.NotNil(t, be)
	defer be.Close()

	// Backup, insert more keys, and backup again.
	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	err = insertKeys(db, 10, 20)
	require.NoError(t, err, "Error inserting keys")

	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	// Get and print backup info.
	backup_info := be.GetInfo()
	backups := backup_info.GetCount()
	// fmt.Printf("Num. backups: %d\n", backups)
	for i := 0; i < backups; i++ {
		// fmt.Printf("--------------\nBackup: %d\n", i)
		// fmt.Printf("Backup ID: %d\n", backup_info.GetBackupId(i))
		// fmt.Printf("Num. files: %d\n", backup_info.GetNumFiles(i))
		// fmt.Printf("Size: %d\n", backup_info.GetSize(i))

		err = be.VerifyBackup(uint32(backup_info.GetBackupId(i)))
		require.NoError(t, err, "Error verifying backup.")
	}

	_ = db.Close()
}

// Test from:
// https://github.com/facebook/rocksdb/wiki/How-to-backup-RocksDB%3F#restoring-a-backup
func TestRestore(t *testing.T) {
	var err error

	backupDir, err := ioutil.TempDir("", "rocksdb-backup")
	require.NoError(t, err)

	// Create new DB and insert keys
	db, _ := newTestDB(t, "original", nil)
	err = insertKeys(db, 10, 0)
	require.NoError(t, err, "Error inserting keys")

	fo := NewDefaultFlushOptions()
	db.Flush(fo)
	fo.Destroy()

	// Create a backup-engine.
	be, err := OpenBackupEngine(NewDefaultOptions(), backupDir)
	require.NoError(t, err)
	require.NotNil(t, be)

	// Backup DB and close backup-engine.
	err = be.CreateNewBackup(db)
	require.NoError(t, err)
	be.Close()

	// Create a different backup engine pointing to backup dir.
	be2, err := OpenBackupEngine(NewDefaultOptions(), backupDir)
	require.NoError(t, err)
	require.NotNil(t, be2)
	defer be2.Close()

	// Restore backup to a specific path.
	restorePath := "/tmp/restored"
	ro := NewRestoreOptions()
	ro.SetKeepLogFiles(1)
	err = be2.RestoreDBFromLatestBackup(restorePath, restorePath, ro)
	require.NoError(t, err)

	// Create the new DB from restored path.
	db2 := newTestDBfromPath(t, restorePath, nil)
	defer db2.Close()

	// Check keys from restored DB.
	it := db2.NewIterator(NewDefaultReadOptions())
	defer it.Close()
	i := 0
	for it.SeekToFirst(); it.Valid(); it.Next() {
		require.Equal(t, []byte("key"+string(i)), it.Key().Data())
		i++
	}
}
