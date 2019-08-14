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
	"net/http"
	"time"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/consensus"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log2"
	"github.com/bbva/qed/protocol"
)

type ClientApi interface {
	Add(event []byte) (*balloon.Snapshot, error)
	AddBulk(bulk [][]byte) ([]*balloon.Snapshot, error)
	QueryDigestMembershipConsistency(keyDigest hashing.Digest, version uint64) (*balloon.MembershipProof, error)
	QueryMembershipConsistency(event []byte, version uint64) (*balloon.MembershipProof, error)
	QueryDigestMembership(keyDigest hashing.Digest) (*balloon.MembershipProof, error)
	QueryMembership(event []byte) (*balloon.MembershipProof, error)
	QueryConsistency(start, end uint64) (*balloon.IncrementalProof, error)
	ClusterInfo() *consensus.ClusterInfo
	Info() *consensus.NodeInfo
}

// HealthCheckResponse contains the response from HealthCheckHandler.
type HealthCheckResponse struct {
	Version int    `json:"version"`
	Status  string `json:"status"`
}

// NewApiHttp returns a new *http.ServeMux containing the current API handlers.
//	/health-check -> Qed server healthcheck
//	/events -> Add event operation
//	/events/bulk -> Add event bulk operation
//	/proofs/membership -> Membership query using event
//	/proofs/digest-membership -> Membership query using event digest
//	/proofs/incremental -> Incremental query
//	/info -> Qed server information
//	/info/shards -> Qed cluster information
func NewApiHttp(api ClientApi) *http.ServeMux {

	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", AuthHandlerMiddleware(HealthCheckHandler))
	mux.HandleFunc("/events", AuthHandlerMiddleware(Add(api)))
	mux.HandleFunc("/events/bulk", AuthHandlerMiddleware(AddBulk(api)))
	mux.HandleFunc("/proofs/membership", AuthHandlerMiddleware(Membership(api)))
	mux.HandleFunc("/proofs/digest-membership", AuthHandlerMiddleware(DigestMembership(api)))
	mux.HandleFunc("/proofs/incremental", AuthHandlerMiddleware(Incremental(api)))
	mux.HandleFunc("/info", AuthHandlerMiddleware(InfoHandler(api)))
	mux.HandleFunc("/info/shards", AuthHandlerMiddleware(InfoShardsHandler(api)))

	return mux
}

// AuthHandlerMiddleware function is an HTTP handler wrapper that performs
// simple authorization tasks. Currently only checks that Api-Key is present.
//
// If Api-Key is not present, it will raise a `http.StatusUnauthorized` errror.
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

// HealthCheckHandler checks the system status and returns it accordinly.
// The http call it answer is:
//	HEAD /
//
// The following statuses are expected:
//
// If everything is alright, the HTTP response will have a 204 status code
// and no body.
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {

	HealthCheckRequest.Inc()
	defer HealthCheckRequest.Dec()
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
//     "EventDigest":   "5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b",
//     "HistoryDigest": "b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846",
//     "HyperDigest":   "6a050f12acfc22989a7681f901a68ace8a9a3672428f8a877f4d21568123a0cb",
//     "Version": 0
//   }
func Add(api ClientApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		AddRequest.Inc()
		defer AddRequest.Dec()
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
// [
//  {
//    "EventDigest":   "5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b",
//    "HistoryDigest": "b8fdd4b2146fe560f94d7a48f8bb3eaf6938f7de6ac6d05bbe033787d8b71846",
//    "HyperDigest":   "6a050f12acfc22989a7681f901a68ace8a9a3672428f8a877f4d21568123a0cb",
//    "Version": 0
//  },
//  {
//    "EventDigest":   "5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b",
//    "HistoryDigest": "4f95cd9fd828abe86b092e506bbffd4662d9431c5755d68eed1ba5e5156fdb13",
//    "HyperDigest":   "7bd6cee5eb0b92801ed4ce58c54a76907221bb4e056165679977b16487e5f015",
//    "Version": 1
//	},
//	...
// ]
func AddBulk(api ClientApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		AddBulkRequest.Inc()
		defer AddBulkRequest.Dec()
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

// Membership returns the membership proof for a given event
// The http post url is:
//   POST /proofs/membership
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 200 and the body contains:
// {
// 	"Exists":           true,
// 	"HyperProof":		"<truncated for clarity in docs>"],
// 	"HistoryProof":		"<truncated for clarity in docs>"],
// 	"CurrentVersion":	3,
// 	"QueryVersion": 	3,
//  "ActualVersion":	0,
// 	"KeyDigest":		"5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b"
// }
func Membership(api ClientApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		MembershipRequest.Inc()
		defer MembershipRequest.Dec()
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

// DigestMembership returns the membership proof for a given event digest
// The http post url is:
//   POST /proofs/digest-membership
//
// Differs from Membership in that instead of sending the raw event we query
// with the keyDigest which is the digest of the event.
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 200 and the body contains:
// {
//  "Exists": 			true,
//  "HyperProof":		"<truncated for clarity in docs>"],
//  "HistoryProof":		"<truncated for clarity in docs>"],
//  "CurrentVersion":	3,
//  "QueryVersion": 	3,
//  "ActualVersion":	0,
//  "KeyDigest":		"5beeaf427ee0bfcd1a7b6f63010f2745110cf23ae088b859275cd0aad369561b"
// }
func DigestMembership(api ClientApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		DigestMembershipRequest.Inc()
		defer DigestMembershipRequest.Dec()

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

// Incremental returns an incremental proof for between initial and end events
// The http post url is:
//   POST /proofs/incremental
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 200 and the body contains:
//   {
//     "Start": "2",
//     "End": "8",
//     "AuditPath": ["<truncated for clarity in docs>"]
//   }
func Incremental(api ClientApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		IncrementalRequest.Inc()
		defer IncrementalRequest.Dec()

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

// InfoShardsHandler returns information about QED shards.
// The http post url is:
//   GET /info/shards
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 200 and the body contains:
// {
//  "NodeId":    "node01",
//  "LeaderId":  "node01",
//  "URIScheme": "http",
//  "Shards": {
//    {
// 	    "NodeId":   "node01",
// 		"HTTPAddr": "http://127.0.0.1:8800"
//    },
//    {
//      "NodeId":   "node02",
// 	    "HTTPAddr": "http://127.0.0.1:8801"
//    }
//   }
// }
func InfoShardsHandler(api ClientApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		InfoShardsRequest.Inc()
		defer InfoShardsRequest.Dec()

		var err error
		// Make sure we can only be called with an HTTP GET request.
		w, _, err = GetReqSanitizer(w, r)
		if err != nil {
			return
		}

		var scheme string
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}

		clusterInfo := api.ClusterInfo()
		if clusterInfo.LeaderId == "" {
			http.Error(w, "Leader not found", http.StatusServiceUnavailable)
			return
		}

		if len(clusterInfo.Nodes) == 0 {
			http.Error(w, "Nodes not found", http.StatusServiceUnavailable)
			return
		}

		nodeInfo := api.Info()
		shardDetails := make(map[string]protocol.ShardDetail)

		for _, node := range clusterInfo.Nodes {
			shardDetails[node.NodeId] = protocol.ShardDetail{
				NodeId:   node.NodeId,
				HTTPAddr: node.HttpAddr,
			}
		}

		shards := &protocol.Shards{
			NodeId:    nodeInfo.NodeId,
			LeaderId:  clusterInfo.LeaderId,
			URIScheme: protocol.Scheme(scheme),
			Shards:    shardDetails,
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

// InfoHandler returns information about the QED server.
// The http post url is:
//   GET /info
//
// The following statuses are expected:
// If everything is alright, the HTTP status is 200 and the body contains:
// {
//  "APIKey": 				"my-key",
//  "NodeID": 				"server0",
//  "HTTPAddr": 			"127.0.0.1:8800",
//  "RaftAddr": 			"127.0.0.1:8500",
//  "MgmtAddr": 			"127.0.0.1:8700",
//  "MetricsAddr": 			"127.0.0.1:8600",
//  "RaftJoinAddr": 		"[]",
//  "DBPath": 				"/var/tmp/db",
//  "RaftPath": 			"/var/tmp/raft",
//  "GossipAddr":		 	"127.0.0.1:8400",
//  "GossipJoinAddr": 		"[]",
//  "PrivateKeyPath": 		"/var/tmp",
//  "EnableTLS": 			false,
//  "EnableProfiling": 		false,
//  "ProfilingAddr": 		"127.0.0.1:6060",
//  "SSLCertificate": 		"/var/tmp/certs/my-cert",
//  "SSLCertificateKey": 	"/var/tmp/certs",
// }
func InfoHandler(api ClientApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		InfoRequest.Inc()
		defer InfoRequest.Dec()
		var err error

		// Make sure we can only be called with an HTTP GET request.
		w, _, err = GetReqSanitizer(w, r)
		if err != nil {
			return
		}

		out, err := json.Marshal(api.Info())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(out)
		return

	}
}

// PostReqSanitizer function checks that certain request info exists and it is correct.
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

// GetReqSanitizer function checks that certain request info exists and it is correct.
func GetReqSanitizer(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, *http.Request, error) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return w, r, errors.New("Method not allowed.")
	}

	return w, r, nil
}

// LogHandler Logs the Http Status for a request into fileHandler and returns a
// httphandler function which is a wrapper to log the requests.
func LogHandler(handle http.Handler, logger log2.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		start := time.Now()
		writer := statusWriter{w, 0, 0}
		handle.ServeHTTP(&writer, request)
		latency := time.Now().Sub(start)

		logger.Debugf("Request: lat %d %+v", latency, request)
		if writer.status >= 400 && writer.status < 500 {
			logger.Infof("Bad Request: %d %+v", latency, request)
		}
		if writer.status >= 500 {
			logger.Infof("Server error: %d %+v", latency, request)
		}
	}
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
