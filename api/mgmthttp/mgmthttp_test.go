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

package mgmthttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/storage"
)

type fakeRaftNode struct {
}

func (b fakeRaftNode) CreateBackup() error {
	return nil
}

func (b fakeRaftNode) ListBackups() []*storage.BackupInfo {
	bi := make([]*storage.BackupInfo, 1)
	info := &storage.BackupInfo{
		ID: 1,
	}
	bi[0] = info
	return bi
}

func (b fakeRaftNode) DeleteBackup(backupID uint32) error {
	return nil
}

func TestCreateBackup(t *testing.T) {
	req, err := http.NewRequest("POST", "/backup", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := ManageBackup(fakeRaftNode{})

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestListBackups(t *testing.T) {
	req, err := http.NewRequest("GET", "/backups", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := ListBackups(fakeRaftNode{})

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	backupsInfo := make([]*protocol.BackupInfo, 1)
	_ = json.Unmarshal([]byte(rr.Body.String()), backupsInfo)
	require.NotNil(t, backupsInfo, "Backups info list must not be empty.")
	require.True(t, len(backupsInfo) == 1, "Backups info list must have 1 element.")
}

func TestDeleteBackup(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/backup?backupID=1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := ManageBackup(fakeRaftNode{})

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNoContent)
	}
}
