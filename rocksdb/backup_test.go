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

func setupDB(t *testing.T) (*DB, string, *BackupEngine, string) {
	var err error

	backupDir, err := ioutil.TempDir("/var/tmp", "backup")
	require.NoError(t, err)
	err = os.RemoveAll(backupDir)
	require.NoError(t, err)

	// Create new DB and insert keys
	db, dbPath := newTestDB(t, "original", nil)
	err = insertKeys(db, 10, 0)
	require.NoError(t, err, "Error inserting keys")
	_ = db.Flush(NewDefaultFlushOptions())

	// Create a backup engine
	be, err := OpenBackupEngine(NewDefaultOptions(), backupDir)
	require.NoError(t, err)
	require.NotNil(t, be)

	return db, dbPath, be, backupDir
}

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

func cleanDirs(dirs ...string) error {
	var err error
	for _, dir := range dirs {
		err = os.RemoveAll(dir)
		if err != nil {
			return err
		}
	}
	return nil
}

// Test from:
// https://github.com/facebook/rocksdb/wiki/How-to-backup-RocksDB%3F#creating-and-verifying-a-backup
func TestBackup(t *testing.T) {
	var err error

	db, dbPath, be, backupDir := setupDB(t)
	defer db.Close()
	defer be.Close()

	// Backup, insert more keys, and backup again.
	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	err = insertKeys(db, 10, 20)
	require.NoError(t, err, "Error inserting keys")

	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	// Verify backup integrity.
	backup_info := be.GetInfo()
	backups := backup_info.GetCount()
	for i := 0; i < backups; i++ {
		err = be.VerifyBackup(uint32(backup_info.GetBackupId(i)))
		require.NoError(t, err, "Error verifying backup.")
	}

	// On success, clean dirs.
	err = cleanDirs(dbPath, backupDir)
	require.NoError(t, err, "Error cleaning directories")
}

func TestBackupWithMetadata(t *testing.T) {
	var err error

	db, dbPath, be, backupDir := setupDB(t)
	defer db.Close()
	defer be.Close()

	// Backup DB with certain metadata.
	metadata := "foo=bar"
	err = be.CreateNewBackupWithMetadata(db, metadata)
	require.NoError(t, err)

	// Verify backup integrity, and check backup metadata.
	backup_info := be.GetInfo()
	backups := backup_info.GetCount()
	for i := 0; i < backups; i++ {
		err = be.VerifyBackup(uint32(backup_info.GetBackupId(i)))
		require.NoError(t, err, "Error verifying backup.")
		require.Equal(t, metadata, backup_info.GetAppMetadata(i), "Metadatas don't match")
	}

	// On success, clean dirs.
	err = cleanDirs(dbPath, backupDir)
	require.NoError(t, err, "Error cleaning directories")
}

func TestMetadataInBackupWithoutMetadata(t *testing.T) {
	var err error

	db, dbPath, be, backupDir := setupDB(t)
	defer db.Close()
	defer be.Close()

	// Backup DB WITHOUT metadata.
	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	// Verify backup integrity, and check backup metadata.
	backup_info := be.GetInfo()
	backups := backup_info.GetCount()
	for i := 0; i < backups; i++ {
		err = be.VerifyBackup(uint32(backup_info.GetBackupId(i)))
		require.NoError(t, err, "Error verifying backup.")
		require.Empty(t, backup_info.GetAppMetadata(i), "Metadata should be empty")
	}

	// On success, clean dirs.
	err = cleanDirs(dbPath, backupDir)
	require.NoError(t, err, "Error cleaning directories")
}

// Test from:
// https://github.com/facebook/rocksdb/wiki/How-to-backup-RocksDB%3F#restoring-a-backup
func TestRestore(t *testing.T) {
	var err error

	db, dbPath, be, backupDir := setupDB(t)
	defer db.Close()
	defer be.Close()

	// Backup DB
	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	// Restore backup to a specific path.
	restorePath, _ := ioutil.TempDir("/var/tmp", "rocksdb-restored")
	ro := NewRestoreOptions()
	ro.SetKeepLogFiles(1)
	err = be.RestoreDBFromLatestBackup(restorePath, restorePath, ro)
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

	// On success, clean dirs.
	err = cleanDirs(dbPath, backupDir, restorePath)
	require.NoError(t, err, "Error cleaning directories")
}

func TestBackupAndRestoreInAnEmptyExistingDB(t *testing.T) {
	var err error

	db, dbPath, be, backupDir := setupDB(t)
	defer db.Close()
	defer be.Close()

	// Backup the original DB.
	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	// Restore backup to a specific path.
	// Before that, create an empty DB on this path and close it inmediately.
	restorePath, _ := ioutil.TempDir("/var/tmp", "rocksdb-restored")

	db2 := newTestDBfromPath(t, restorePath, nil)
	db2.Close()

	ro := NewRestoreOptions()
	ro.SetKeepLogFiles(1)
	err = be.RestoreDBFromLatestBackup(restorePath, restorePath, ro)
	require.NoError(t, err)

	// Create a DB on path with restored backup.
	db2 = newTestDBfromPath(t, restorePath, nil)
	defer db2.Close()

	// Check keys from restored DB.
	it := db2.NewIterator(NewDefaultReadOptions())
	defer it.Close()
	i := 0
	for it.SeekToFirst(); it.Valid(); it.Next() {
		require.Equal(t, []byte("key"+string(i)), it.Key().Data())
		i++
	}

	// On success, clean dirs.
	err = cleanDirs(dbPath, backupDir, restorePath)
	require.NoError(t, err, "Error cleaning directories")
}

func TestMultipleBackupsAndRestores(t *testing.T) {
	var err error

	db, dbPath, be, backupDir := setupDB(t)
	defer db.Close()
	defer be.Close()

	// Backup the original DB.
	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	// Restore backup to a specific path.
	restorePath, _ := ioutil.TempDir("/var/tmp", "rocksdb-restored")
	ro := NewRestoreOptions()
	ro.SetKeepLogFiles(1)
	err = be.RestoreDBFromLatestBackup(restorePath, restorePath, ro)
	require.NoError(t, err)

	// Create a DB on path with restored backup.
	db2 := newTestDBfromPath(t, restorePath, nil)

	// Check keys from restored DB and close it.
	it := db2.NewIterator(NewDefaultReadOptions())
	i := 0
	for it.SeekToFirst(); it.Valid(); it.Next() {
		require.Equal(t, []byte("key"+string(i)), it.Key().Data())
		i++
	}
	it.Close()
	db2.Close()

	// Insert more keys on the original DB, and backup it.
	err = insertKeys(db, 10, 10)
	require.NoError(t, err)
	err = be.CreateNewBackup(db)
	require.NoError(t, err)

	// Restore from the latests backup to the same "restore path" as before.
	err = be.RestoreDBFromLatestBackup(restorePath, restorePath, ro)
	require.NoError(t, err)

	// Open DB2 again from restored path, and check keys.
	db2, err = OpenDB(restorePath, NewDefaultOptions())
	defer db2.Close()
	require.NoError(t, err)

	it = db2.NewIterator(NewDefaultReadOptions())
	defer it.Close()

	i = 0
	for it.SeekToFirst(); it.Valid(); it.Next() {
		require.Equal(t, []byte("key"+string(i)), it.Key().Data())
		i++
	}

	// On success, clean dirs.
	err = cleanDirs(dbPath, backupDir, restorePath)
	require.NoError(t, err, "Error cleaning directories")
}
