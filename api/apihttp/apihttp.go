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

// Package apihttp implements the HTTP API public interface.
package apihttp

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/bbva/qed/balloon"
)

// Struct HealthCheckResponse contains the response from HealthCheckHandler.
type HealthCheckResponse struct {
	Version int    `json:"version"`
	Status  string `json:"status"`
}

// This handler checks the system status and returns it accordinly.
// The http call it answer is:
//	GET /health-check
//
// The following statuses are expected:
//
// If everything is allright, the HTTP status is 200 and the body contains:
//	 {"version": "0", "status":"ok"}
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	result := HealthCheckResponse{
		Version: 0,
		Status:  "ok",
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		panic(err)
	}

	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	out := new(bytes.Buffer)
	json.Compact(out, resultJson)

	w.Write(out.Bytes())
}

// This handler posts an event into the system:
// The http post url is:
//   POST /events
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 201 and the body contains:
//   {
//     "HyperDigest": "mHzXvSE/j7eFmNObvC7PdtQTmd4W0q/FPHmiYEjL0eM=",
//     "HistoryDigest": "Kpbn+7P4XrZi2hKpdhA7freUicZdUsU6GqmUk0vDJ8A=",
//     "Version": 1,
//     "Event": "VGhpcyBpcyBteSBmaXJzdCBldmVudA=="
//   }
func Add(balloon balloon.Balloon) http.HandlerFunc {

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

		var event Event
		err := json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Wait for the response
		response := <-balloon.Add(event.Event)
		out, err := json.Marshal(ToSnapshot(response, event.Event))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write(out)
		return

	}
}

// This handler returns a membershipProof from the system
// The http post url is:
//   POST /proofs/membership
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 201 and the body contains:
//   {
//     "key": "TG9yZW0gaXBzdW0gZGF0dW0gbm9uIGNvcnJ1cHR1bSBlc3QK",
//     "keyDigest": "NDRkMmY3MjEzYjlhMTI4ZWRhZjQzNWFhNjcyMzUxMGE0YTRhOGY5OWEzOWNiYTVhN2FhMWI5OWEwYTlkYzE2NCAgLQo=",
//     "isMember": "true",
//     "proofs": ["<truncated for clarity in docs>"],
//     "queryVersion": "1",
//     "actualVersion": "2",
//   }
func Membership(balloon balloon.Balloon) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Make sure we can only be called with an HTTP POST request.
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var query MembershipQuery
		err := json.NewDecoder(r.Body).Decode(&query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Wait for the response
		proof := <-balloon.GenMembershipProof(query.Key, query.Version)

		out, err := json.Marshal(ToMembershipProof(query.Key, proof))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(out)
		return

	}
}

// AuthHandlerMiddleware function is an HTTP handler wrapper that validates our requests
func AuthHandlerMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if Api-Key header is empty
		if r.Header.Get("Api-Key") == "" {
			http.Error(w, "Missing Api-Key header", http.StatusUnauthorized)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

// NewServer returns a new *http.ServeMux containing all the API handlers
// already configured
func NewApiHttp(balloon balloon.Balloon) *http.ServeMux {

	api := http.NewServeMux()
	api.HandleFunc("/health-check", AuthHandlerMiddleware(HealthCheckHandler))
	api.HandleFunc("/events", AuthHandlerMiddleware(Add(balloon)))
	api.HandleFunc("/proofs/membership", AuthHandlerMiddleware(Membership(balloon)))

	return api
}
