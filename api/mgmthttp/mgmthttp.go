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

// Package mgmthttp implements the Raft management HTTP API public interface.
package mgmthttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/raftwal"
)

// NewMgmtHttp will return a mux server with endpoints to manage different
// QED log service features: DDBB backups, Raft membership,...
func NewMgmtHttp(balloon raftwal.RaftBalloonApi) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/join", joinHandle(balloon))
	mux.HandleFunc("/backup", apihttp.AuthHandlerMiddleware(ManageBackup(balloon)))
	mux.HandleFunc("/backups", apihttp.AuthHandlerMiddleware(ListBackups(balloon)))
	return mux
}

func joinHandle(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		// Make sure we can only be called with an HTTP POST request.
		w, r, err = apihttp.PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		body := make(map[string]interface{})

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(body) != 3 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		remoteAddr, ok := body["addr"].(string)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		nodeID, ok := body["id"].(string)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		m, ok := body["metadata"].(map[string]interface{})
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// TO IMPROVE: use map[string]interface{} for nested metadata.
		metadata := make(map[string]string)
		for k, v := range m {
			metadata[k] = v.(string)
		}

		if err := api.Join(nodeID, remoteAddr, metadata); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func ManageBackup(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "DELETE":
			DeleteBackup(api, w, r)
		case "POST":
			CreateBackup(api, w, r)
		default:
			w.Header().Set("Allow", "POST, DELETE")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	}
}

// CreateBackup creates a backup of the RocksDB data up to now:
// The http post url is:
//   POST /backup
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 200 with an empty body.
func CreateBackup(api raftwal.RaftBalloonApi, w http.ResponseWriter, r *http.Request) {
	// Make sure we can only be called with an HTTP POST request.
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := api.Backup(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListBackups returns a list of backups along with each backup information.
// The http post url is:
//   GET /backups
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 204 and the body contains:
// [
//  {
//    "ID":   1,
//    "Timestamp": 1523653256854,
//    "Size":   16786",
//    "NumFiles": 4,
//	  "Metadata": "foo"
//  },
//  {
//    "ID":   2,
//    "Timestamp": 1523653256999,
//    "Size":   21786,
//    "NumFiles": 4,
//	  "Metadata": "bar"
//	},
//	...
// ]
func ListBackups(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		// Make sure we can only be called with an HTTP GET request.
		w, _, err = apihttp.GetReqSanitizer(w, r)
		if err != nil {
			return
		}

		backups := api.ListBackups()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out, err := json.Marshal(backups)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
	}
}

// DeleteBackup deletes a certain backup (given its ID) from the system:
// The http post url is:
//   DELETE /backup?backupID=<id>
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 204 with an empty body.
func DeleteBackup(api raftwal.RaftBalloonApi, w http.ResponseWriter, r *http.Request) {

	if r.Method != "DELETE" {
		w.Header().Set("Allow", "DELETE")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	b := r.URL.Query()["backupID"]
	backupID, err := strconv.ParseUint(b[0], 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := api.DeleteBackup(uint32(backupID)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
