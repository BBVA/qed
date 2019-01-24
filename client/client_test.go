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

package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/stretchr/testify/assert"
)

var (
	client *HTTPClient
	mux    *http.ServeMux
	server *httptest.Server
)

func init() {
	log.SetLogger("client-test", "info")
}

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	mux.HandleFunc("/info/shards", infoHandler(server.URL))

	client = NewHTTPClient(Config{
		Cluster:  QEDCluster{Endpoints: []string{server.URL}, Leader: server.URL},
		APIKey:   "my-awesome-api-key",
		Insecure: false,
	})
	return func() {
		server.Close()
	}
}

func TestAddSuccess(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	event := "Hello world!"
	snap := &protocol.Snapshot{
		HistoryDigest: []byte("history"),
		HyperDigest:   []byte("hyper"),
		Version:       0,
		EventDigest:   []byte(event),
	}

	result, _ := json.Marshal(snap)
	mux.HandleFunc("/events", okHandler(result))

	snapshot, err := client.Add(event)
	assert.NoError(t, err)
	assert.Equal(t, snap, snapshot, "The snapshots should match")

}

func TestAddWithServerFailure(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	event := "Hello world!"
	mux.HandleFunc("/events", serverErrorHandler())

	_, err := client.Add(event)
	assert.Error(t, err)

}

func TestMembership(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	event := "Hello world!"
	version := uint64(0)
	fakeResult := &protocol.MembershipResult{
		Key:            []byte(event),
		KeyDigest:      []byte("digest"),
		Exists:         true,
		Hyper:          make(visitor.AuditPath),
		History:        make(visitor.AuditPath),
		CurrentVersion: version,
		QueryVersion:   version,
		ActualVersion:  version,
	}
	resultJSON, _ := json.Marshal(fakeResult)
	mux.HandleFunc("/proofs/membership", okHandler(resultJSON))

	result, err := client.Membership([]byte(event), version)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The results should match")
}

func TestDigestMembership(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	event := "Hello world!"
	version := uint64(0)
	fakeResult := &protocol.MembershipResult{
		Key:            []byte(event),
		KeyDigest:      []byte("digest"),
		Exists:         true,
		Hyper:          make(visitor.AuditPath),
		History:        make(visitor.AuditPath),
		CurrentVersion: version,
		QueryVersion:   version,
		ActualVersion:  version,
	}
	resultJSON, _ := json.Marshal(fakeResult)
	mux.HandleFunc("/proofs/digest-membership", okHandler(resultJSON))

	result, err := client.MembershipDigest([]byte("digest"), version)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The results should match")
}

func TestMembershipWithServerFailure(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	event := "Hello world!"
	mux.HandleFunc("/proofs/membership", serverErrorHandler())

	_, err := client.Membership([]byte(event), 0)
	assert.Error(t, err)
}

func TestIncremental(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	start := uint64(2)
	end := uint64(8)
	fakeResult := &protocol.IncrementalResponse{
		Start:     start,
		End:       end,
		AuditPath: visitor.AuditPath{"0|0": []uint8{0x0}},
	}

	resultJSON, _ := json.Marshal(fakeResult)
	mux.HandleFunc("/proofs/incremental", okHandler(resultJSON))

	result, err := client.Incremental(start, end)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The results should match")
}

func TestIncrementalWithServerFailure(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	mux.HandleFunc("/proofs/incremental", serverErrorHandler())

	_, err := client.Incremental(uint64(2), uint64(8))
	assert.Error(t, err)

}

// TODO implement a test to verify proofs using fake hash function

func okHandler(result []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		out := new(bytes.Buffer)
		_ = json.Compact(out, result)
		_, _ = w.Write(out.Bytes())
	}
}

func serverErrorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func infoHandler(serverURL string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		var md = make(map[string]interface{})
		md["nodeID"] = "node01"
		md["leaderID"] = "node01"
		md["URIScheme"] = "http://"
		md["meta"] = map[string]map[string]string{
			"node01": map[string]string{
				"HTTPAddr": strings.Trim(serverURL, "http://"),
			},
		}

		out, _ := json.Marshal(md)
		_, _ = w.Write(out)
	}
}
