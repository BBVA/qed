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

package tampering

import (
	"encoding/json"
	"net/http"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

type tamperEvent struct {
	Key       []byte
	KeyDigest []byte
	Value     []byte
}

// NewTamperingApi will return a mux server with the endpoint required to
// tamper the server. it's a internal debug implementation. Running a server
// with this enabled will run useless the qed server.
func NewTamperingApi(store storage.DeletableStore, hasher hashing.Hasher) *http.ServeMux {
	api := http.NewServeMux()
	api.HandleFunc("/tamper", apihttp.AuthHandlerMiddleware(http.HandlerFunc(tamperFunc(store, hasher))))
	return api
}

func tamperFunc(store storage.DeletableStore, hasher hashing.Hasher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Make sure we can only be called with an HTTP POST request.
		if !(r.Method == "PATCH" || r.Method == "DELETE") {
			w.Header().Set("Allow", "PATCH, DELETE")
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

		tp.KeyDigest = hasher.Do(tp.Key)

		switch r.Method {
		case "PATCH":
			get, _ := store.Get(storage.IndexPrefix, tp.KeyDigest)
			log.Debugf("Get: %v", get)
			mutations := make([]storage.Mutation, 0)
			mutations = append(mutations, *storage.NewMutation(storage.IndexPrefix, tp.KeyDigest, tp.Value))
			log.Debugf("Tamper: %v", store.Mutate(mutations))

		case "DELETE":
			get, _ := store.Get(storage.IndexPrefix, tp.KeyDigest)
			log.Debugf("Get: %v", get)
			log.Debugf("Delete: %v", store.Delete(storage.IndexPrefix, tp.KeyDigest))

		}

		return
	}
}
