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

package tampering

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/raftwal"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/util"
)

type tamperEvent struct {
	Digest string
	Value  uint64
}

// NewTamperingAPI will return a mux server with the endpoint required to
// tamper the server. it's a internal debug implementation. Running a server
// with this enabled will run useless the qed server.
func NewTamperingAPI(store storage.DeletableStore, balloon raftwal.RaftBalloonApi) *http.ServeMux {
	api := http.NewServeMux()
	api.HandleFunc("/hyper/store", apihttp.AuthHandlerMiddleware(hyperStore(store)))
	api.HandleFunc("/hyper/add", apihttp.AuthHandlerMiddleware(hyperAdd(balloon)))
	api.HandleFunc("/history/store", apihttp.AuthHandlerMiddleware(historyStore(store)))
	return api
}

func hyperAdd(balloon raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Make sure we can only be called with an HTTP POST request.
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}

		var tp tamperEvent
		err := json.NewDecoder(r.Body).Decode(&tp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		digest, _ := hex.DecodeString(tp.Digest)

		// Wait for the response
		response, err := balloon.TamperHyper(digest, tp.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		snapshot := &protocol.Snapshot{
			response.HistoryDigest,
			response.HyperDigest,
			response.Version,
			response.EventDigest,
		}

		out, err := json.Marshal(snapshot)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(out)

		return
	}
}

func historyStore(store storage.DeletableStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}

		var tp tamperEvent
		err = json.NewDecoder(r.Body).Decode(&tp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		digest, _ := hex.DecodeString(tp.Digest)
		version := util.Uint64AsBytes(tp.Value)

		mutations := make([]*storage.Mutation, 0)

		switch r.Method {
		case "PATCH":
			// Size of the index plus 2 bytes for the height, which is always 0,
			// as it is always a leaf what we want to tamper
			bIndex := make([]byte, 10)
			copy(bIndex, version)
			mutation := &storage.Mutation{storage.HistoryCachePrefix, bIndex, digest}
			mutations = append(mutations, mutation)
			err := store.Mutate(mutations)
			log.Debugf("tamper_api: patch store mutations %+v --> %v", mutations, err)

		case "DELETE":
			// Size of the index plus 2 bytes for the height, which is always 0,
			// as it is always a leaf what we want to tamper
			bIndex := make([]byte, 10)
			copy(bIndex, version)
			err := store.Delete(storage.HistoryCachePrefix, bIndex)
			if err != nil {
				http.Error(w, "Error deleting hyper digest", http.StatusUnprocessableEntity)
				return
			}
			log.Debugf("tamper_api: deleted mutations %+v -->: %v", mutations, err)
		}
		return
	}

}

func hyperStore(store storage.DeletableStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error

		if r.Body == nil {
			http.Error(w, "Please send a request body", http.StatusBadRequest)
			return
		}

		var tp tamperEvent
		err = json.NewDecoder(r.Body).Decode(&tp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		digest, _ := hex.DecodeString(tp.Digest)
		version := util.Uint64AsBytes(tp.Value)

		mutations := make([]*storage.Mutation, 0)

		switch r.Method {
		case "PATCH":
			mutation := &storage.Mutation{storage.IndexPrefix, digest, version}
			mutations = append(mutations, mutation)
			err := store.Mutate(mutations)
			log.Debugf("tamper_api: patch store mutations %+v --> %v", mutations, err)

		case "DELETE":
			err := store.Delete(storage.IndexPrefix, digest)
			if err != nil {
				http.Error(w, "Error deleting hyper digest", http.StatusUnprocessableEntity)
				return
			}
			log.Debugf("tamper_api: deleted mutations %+v -->: %v", mutations, err)
		}
		return
	}
}
