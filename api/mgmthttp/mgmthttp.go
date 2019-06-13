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

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/raftwal"
)

// NewMgmtHttp will return a mux server with the endpoint required to
// join the raft cluster.
func NewMgmtHttp(raftBalloon raftwal.RaftBalloonApi) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/join", joinHandle(raftBalloon))
	return mux
}

func joinHandle(raftBalloon raftwal.RaftBalloonApi) http.HandlerFunc {
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

		if err := raftBalloon.Join(nodeID, remoteAddr, metadata); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
