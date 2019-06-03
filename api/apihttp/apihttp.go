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

// Package apihttp implements the HTTP API public interface.
package apihttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/raftwal"
)

// HealthCheckResponse contains the response from HealthCheckHandler.
type HealthCheckResponse struct {
	Version int    `json:"version"`
	Status  string `json:"status"`
}

// HealthCheckHandler checks the system status and returns it accordinly.
// The http call it answer is:
//	HEAD /
//
// The following statuses are expected:
//
// If everything is alright, the HTTP response will have a 204 status code
// and no body.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	// Make sure we can only be called with an HTTP POST request.
	if r.Method != "HEAD" {
		w.Header().Set("Allow", "HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// A very simple health check.
	w.WriteHeader(http.StatusNoContent)

}

// Add posts an event into the system:
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
func Add(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		// Make sure we can only be called with an HTTP POST request.
		w, r, err = PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		var event protocol.Event
		err = json.NewDecoder(r.Body).Decode(&event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Wait for the response
		response, err := api.Add(event.Event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			return
		}

		snapshot := protocol.Snapshot(*response)

		out, err := json.Marshal(&snapshot)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(out)

		return
	}
}

// AddBulk posts a bulk of events into the system:
// The http post url is:
//   POST /events/bulk
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 201 and the body contains:
//   [
//		{
//			"HyperDigest": "mHzXvSE/j7eFmNObvC7PdtQTmd4W0q/FPHmiYEjL0eM=",
//			"HistoryDigest": "Kpbn+7P4XrZi2hKpdhA7freUicZdUsU6GqmUk0vDJ8A=",
//			"Version": 1,
//			"Event": "VGhpcyBpcyBteSBmaXJzdCBldmVudA=="
//		},
//		{
//			"HyperDigest": "mHzXvSE/j7eFmNObvC7PdtQTmd4W0q/FPHmiYEjL0eM=",
//			"HistoryDigest": "reUicZdUsU6GqmUk0vDJ8A=Kpbn+7P4XrZi2hKpdhA7f",
//			"Version": 2,
//			"Event": "pcyBteSBmaXJzdCBldmVudAVGhpcyB=="
//		},
//		...
//	]
func AddBulk(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		// Make sure we can only be called with an HTTP POST request.
		w, r, err = PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		var eventBulk protocol.EventsBulk
		err = json.NewDecoder(r.Body).Decode(&eventBulk)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Wait for the response
		snapshotBulk, err := api.AddBulk(eventBulk.Events)

		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			return
		}

		out, err := json.Marshal(snapshotBulk)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(out)

		return
	}
}

// Membership returns a membershipProof from the system
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
func Membership(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var proof *balloon.MembershipProof
		var err error

		// Make sure we can only be called with an HTTP POST request.
		w, r, err = PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		var query protocol.MembershipQuery
		err = json.NewDecoder(r.Body).Decode(&query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if query.Version == nil {
			// Wait for the response
			proof, err = api.QueryMembership(query.Key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusPreconditionFailed)
				return
			}
		} else {

			// Wait for the response
			proof, err = api.QueryMembershipConsistency(query.Key, *query.Version)
			if err != nil {
				http.Error(w, err.Error(), http.StatusPreconditionFailed)
				return
			}
		}
		out, err := json.Marshal(protocol.ToMembershipResult(query.Key, proof))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
		return

	}
}

// DigestMembership returns a membershipProof from the system
// The http post url is:
//   POST /proofs/digest-membership
//
// Differs from Membership in that instead of sending the raw event we query
// with the keyDigest which is the digest of the event.
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
func DigestMembership(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var proof *balloon.MembershipProof
		var err error

		// Make sure we can only be called with an HTTP POST request.
		w, r, err = PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		var query protocol.MembershipDigest
		err = json.NewDecoder(r.Body).Decode(&query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if query.Version == nil {
			// Wait for the response
			proof, err = api.QueryDigestMembership(query.KeyDigest)
			if err != nil {
				http.Error(w, err.Error(), http.StatusPreconditionFailed)
				return
			}
		} else {

			// Wait for the response
			proof, err = api.QueryDigestMembershipConsistency(query.KeyDigest, *query.Version)
			if err != nil {
				http.Error(w, err.Error(), http.StatusPreconditionFailed)
				return
			}
		}

		out, err := json.Marshal(protocol.ToMembershipResult(nil, proof))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
		return

	}
}

// Incremental returns an incrementalProof from the system
// The http post url is:
//   POST /proofs/incremental
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 201 and the body contains:
//   {
//     "start": "2",
//     "end": "8",
//     "auditPath": ["<truncated for clarity in docs>"]
//   }
func Incremental(api raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		// Make sure we can only be called with an HTTP POST request.
		w, r, err = PostReqSanitizer(w, r)
		if err != nil {
			return
		}

		var request protocol.IncrementalRequest
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Wait for the response
		proof, err := api.QueryConsistency(request.Start, request.End)
		if err != nil {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			return
		}

		out, err := json.Marshal(protocol.ToIncrementalResponse(proof))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
		return

	}
}

// AuthHandlerMiddleware function is an HTTP handler wrapper that performs
// simple authorization tasks. Currently only checks that Api-Key it's present.
//
// If not present will raise a `http.StatusUnauthorized` errror.
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

// NewApiHttp returns a new *http.ServeMux containing the current API handlers.
//	/health-check -> HealthCheckHandler
//	/events -> Add
//	/proofs/membership -> Membership
func NewApiHttp(balloon raftwal.RaftBalloonApi) *http.ServeMux {

	api := http.NewServeMux()
	api.HandleFunc("/healthcheck", AuthHandlerMiddleware(HealthCheckHandler))
	api.HandleFunc("/events", AuthHandlerMiddleware(Add(balloon)))
	api.HandleFunc("/events/bulk", AuthHandlerMiddleware(AddBulk(balloon)))
	api.HandleFunc("/proofs/membership", AuthHandlerMiddleware(Membership(balloon)))
	api.HandleFunc("/proofs/digest-membership", AuthHandlerMiddleware(DigestMembership(balloon)))
	api.HandleFunc("/proofs/incremental", AuthHandlerMiddleware(Incremental(balloon)))
	api.HandleFunc("/info/shards", AuthHandlerMiddleware(InfoShardsHandler(balloon)))

	return api
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

// LogHandler Logs the Http Status for a request into fileHandler and returns a
// httphandler function which is a wrapper to log the requests.
func LogHandler(handle http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		start := time.Now()
		writer := statusWriter{w, 0, 0}
		handle.ServeHTTP(&writer, request)
		latency := time.Now().Sub(start)

		log.Debugf("Request: lat %d %+v", latency, request)
		if writer.status >= 400 && writer.status < 500 {
			log.Infof("Bad Request: %d %+v", latency, request)
		}
		if writer.status >= 500 {
			log.Infof("Server error: %d %+v", latency, request)
		}
	}
}

func InfoShardsHandler(balloon raftwal.RaftBalloonApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.Header().Set("Allow", "GET")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var scheme string
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}

		info := balloon.Info()
		details := make(map[string]protocol.ShardDetail)
		for k, v := range info["meta"].(map[string]map[string]string) {
			fmt.Println(k, v)
			details[k] = protocol.ShardDetail{
				NodeId:   k,
				HTTPAddr: v["HTTPAddr"],
			}
		}

		shards := &protocol.Shards{
			NodeId:    info["nodeID"].(string),
			LeaderId:  info["leaderID"].(string),
			URIScheme: protocol.Scheme(scheme),
			Shards:    details,
		}

		out, err := json.Marshal(shards)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
	}
}

func PostReqSanitizer(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return w, r, errors.New("Method not allowed.")
	}

	if r.Body == nil {
		http.Error(w, "Please send a request body", http.StatusBadRequest)
		return w, r, errors.New("Bad request: nil body.")
	}

	return w, r, nil
}
