/*
   copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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

package e2e

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/rocksdb"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/spec"
)

func TestRestoreFromBackup(t *testing.T) {
	before, after := newServerSetup(0, false)
	let, report := spec.New()
	defer func() { t.Logf(report()) }()
	// log.SetLogger("", log.DEBUG)

	event := rand.RandomString(10)
	var snapshot *protocol.Snapshot
	var client *client.HTTPClient

	var err error
	serverPath, err := before()
	spec.NoError(t, err, "Error starting server")
	backupPath := serverPath + "/db/backups"
	backupTempPath := "/var/tmp/e2e-backup"

	let(t, "Start a QED Log server, add one event, create a backup, and stop server.", func(t *testing.T) {
		var err error

		client, err = newQedClient(0)
		spec.NoError(t, err, "Error creating qed client")
		defer func() { client.Close() }()

		let(t, "add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			spec.NoError(t, err, "Error adding event")
		})

		let(t, "create backup", func(t *testing.T) {
			_, err = doReq("POST", "http://127.0.0.1:8700/backup", "APIKey", true, nil)
			spec.NoError(t, err, "Error creating backup.")
		})
	})

	let(t, "Copy backup to the new QED server placement.", func(t *testing.T) {
		var err error

		let(t, "move backup folder to a temporary path.", func(t *testing.T) {
			err = copyDirectory(backupPath, backupTempPath)
			spec.NoError(t, err, "Error moving backup folder.")
		})

		let(t, "clean former QED layout.", func(t *testing.T) {
			err = os.RemoveAll(serverPath)
			spec.NoError(t, err, "Error cleaning layout.")
		})
	})

	err = after()
	spec.NoError(t, err, "Error stoping server")

	let(t, "Restore backup, start a new server, and check event membership.", func(t *testing.T) {
		var err error

		client, err = newQedClient(0)
		spec.NoError(t, err, "Error creating qed client")
		defer func() { client.Close() }()

		let(t, "create new QED backup layout from scratch.", func(t *testing.T) {
			err = os.MkdirAll(serverPath, os.ModePerm)
			spec.NoError(t, err, "Error creating layout.")
		})

		let(t, "move backup folder back.", func(t *testing.T) {
			err = copyDirectory(backupTempPath, backupPath)
			spec.NoError(t, err, "Error restoring backup folder.")
		})

		let(t, "restore backup", func(t *testing.T) {
			bo := rocksdb.NewDefaultOptions()
			be, err := rocksdb.OpenBackupEngine(bo, backupPath)
			spec.NoError(t, err, "Error creating backup engine.")

			ro := rocksdb.NewRestoreOptions()
			defer ro.Destroy()
			err = be.RestoreDBFromLatestBackup(serverPath+"/db", serverPath+"/db", ro)
			spec.NoError(t, err, "Error restoring from latest backup.")
		})

		let(t, "start server", func(t *testing.T) {
			_, err = before()
			spec.NoError(t, err, "Error starting server")
		})

		let(t, "get event membership proof.", func(t *testing.T) {
			proof, err := client.Membership([]byte(event), &snapshot.Version)
			spec.NoError(t, err, "Error getting membership proof")

			spec.True(t, proof.Exists, "The queried key should be a member")
			spec.Equal(t, proof.QueryVersion, snapshot.Version, "The query version doest't match the queried one")
			spec.Equal(t, proof.ActualVersion, snapshot.Version, "The actual version should match the queried one")
			spec.Equal(t, proof.CurrentVersion, snapshot.Version, "The current version should match the queried one")
			spec.False(t, len(proof.KeyDigest) == 0, "The key digest cannot be empty")
			spec.NotNil(t, proof.HyperProof, "The hyper proof cannot be empty")
			spec.False(t, proof.ActualVersion > 0 && proof.HistoryProof == nil, "The history proof cannot be empty when version is greater than 0")
		})

	})

	err = after()
	spec.NoError(t, err, "Error stoping server")

	// Cleanup dirs.
	_ = os.RemoveAll(serverPath)
	_ = os.RemoveAll(backupTempPath)
}

/*
	COPY DIRECTORY UTILS
*/

func copyDirectory(scrDir, dest string) error {
	entries, err := ioutil.ReadDir(scrDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(scrDir, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", sourcePath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirectory(sourcePath, destPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := copySymLink(sourcePath, destPath); err != nil {
				return err
			}
		default:
			if err := copy(sourcePath, destPath); err != nil {
				return err
			}
		}

		if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		isSymlink := entry.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func copy(srcFile, dstFile string) error {
	out, err := os.Create(dstFile)
	defer out.Close()
	if err != nil {
		return err
	}

	in, err := os.Open(srcFile)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func createIfNotExists(dir string, perm os.FileMode) error {
	if exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func copySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}
