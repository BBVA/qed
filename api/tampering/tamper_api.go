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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/util"
)

type tamperEvent struct {
	Digest string
	Value  string
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

		digest, _ := hex.DecodeString(tp.Digest)
		value, _ := hex.DecodeString(tp.Value)

		switch r.Method {
		case "PATCH":
			index, err := store.Get(storage.IndexPrefix, digest)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s: %X", err.Error(), index), http.StatusUnprocessableEntity)
				return
			}

			bIndex := make([]byte, 10) // Size of the index plus 2 bytes for the height
			copy(bIndex, index.Value)
			copy(bIndex[len(index.Value):], util.Uint16AsBytes(uint16(0)))

			mutations := []*storage.Mutation{{storage.HistoryCachePrefix, bIndex, value}}

			log.Debugf("Tamper: %v", store.Mutate(mutations))

		case "DELETE":
			_, err = store.Get(storage.IndexPrefix, digest)
			if err != nil {
				http.Error(w, fmt.Sprintf("%s: %X", err.Error(), digest), http.StatusUnprocessableEntity)
				return
			}

			log.Debugf("Delete: %v", store.Delete(storage.IndexPrefix, digest))

		}

		return
	}
}
