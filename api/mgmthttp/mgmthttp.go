/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

	"github.com/bbva/qed/raftwal"
)

// NewMgmtHttp will return a mux server with the endpoint required to
// tamper the server. it's a internal debug implementation. Running a server
// with this enabled will run useless the qed server.
func NewMgmtHttp(raftBalloon raftwal.RaftBalloonApi) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/join", joinHandle(raftBalloon))
	return mux
}

func joinHandle(raftBalloon raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := map[string]string{}

		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(m) != 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		remoteAddr, ok := m["addr"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		nodeID, ok := m["id"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err := raftBalloon.Join(nodeID, remoteAddr); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}
}
