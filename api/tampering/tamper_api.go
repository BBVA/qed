package tampering

import (
	"encoding/json"
	"net/http"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/balloon/hashing"
	"github.com/bbva/qed/balloon/hyper/storage"
	"github.com/bbva/qed/log"
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

		tp.KeyDigest = hasher(tp.Key)

		switch r.Method {
		case "PATCH":
			get, _ := store.Get(tp.KeyDigest)
			log.Debugf("Get: %v", get)
			log.Debugf("Tamper: %v", store.Add(tp.KeyDigest, tp.Value))

		case "DELETE":
			get, _ := store.Get(tp.KeyDigest)
			log.Debugf("Get: %v", get)
			log.Debugf("Delete: %v", store.Delete(tp.KeyDigest))

		}

		return
	}
}
