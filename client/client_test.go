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

	"github.com/bbva/qed/hashing"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/stretchr/testify/assert"
)

func init() {
	log.SetLogger("client-test", "info")
}

func setupServer(input []byte) (string, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc("/info/shards", infoHandler(server.URL))
	mux.HandleFunc("/events", defaultHandler(input))
	mux.HandleFunc("/proofs/membership", defaultHandler(input))
	mux.HandleFunc("/proofs/incremental", defaultHandler(input))
	mux.HandleFunc("/proofs/digest-membership", defaultHandler(input))

	return server.URL, func() {
		server.Close()
	}
}

func setupClient(urls []string) *HTTPClient {
	return NewHTTPClient(Config{
		Endpoints: urls,
		APIKey:    "my-awesome-api-key",
		Insecure:  false,
	})
}

func TestAddSuccess(t *testing.T) {

	event := "Hello world!"
	snap := &protocol.Snapshot{
		HistoryDigest: []byte("history"),
		HyperDigest:   []byte("hyper"),
		Version:       0,
		EventDigest:   []byte(event),
	}
	input, _ := json.Marshal(snap)

	serverURL, tearDown := setupServer(input)
	defer tearDown()
	client := setupClient([]string{serverURL})

	snapshot, err := client.Add(event)
	assert.NoError(t, err)
	assert.Equal(t, snap, snapshot, "The snapshots should match")
}

func TestAddWithServerFailure(t *testing.T) {
	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient([]string{serverURL})

	event := "Hello world!"
	_, err := client.Add(event)
	assert.Error(t, err)
}

func TestMembership(t *testing.T) {
	event := "Hello world!"
	version := uint64(0)
	fakeResult := &protocol.MembershipResult{
		Key:            []byte(event),
		KeyDigest:      []byte("digest"),
		Exists:         true,
		Hyper:          make(map[string]hashing.Digest),
		History:        make(map[string]hashing.Digest),
		CurrentVersion: version,
		QueryVersion:   version,
		ActualVersion:  version,
	}
	inputJSON, _ := json.Marshal(fakeResult)

	serverURL, tearDown := setupServer(inputJSON)
	defer tearDown()
	client := setupClient([]string{serverURL})

	result, err := client.Membership([]byte(event), version)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The inputs should match")
}

func TestDigestMembership(t *testing.T) {

	event := "Hello world!"
	version := uint64(0)
	fakeResult := &protocol.MembershipResult{
		Key:            []byte(event),
		KeyDigest:      []byte("digest"),
		Exists:         true,
		Hyper:          make(map[string]hashing.Digest),
		History:        make(map[string]hashing.Digest),
		CurrentVersion: version,
		QueryVersion:   version,
		ActualVersion:  version,
	}
	inputJSON, _ := json.Marshal(fakeResult)

	serverURL, tearDown := setupServer(inputJSON)
	defer tearDown()
	client := setupClient([]string{serverURL})

	result, err := client.MembershipDigest([]byte("digest"), version)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The results should match")
}

func TestMembershipWithServerFailure(t *testing.T) {
	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient([]string{serverURL})

	event := "Hello world!"

	_, err := client.Membership([]byte(event), 0)
	assert.Error(t, err)
}

func TestIncremental(t *testing.T) {

	start := uint64(2)
	end := uint64(8)
	fakeResult := &protocol.IncrementalResponse{
		Start:     start,
		End:       end,
		AuditPath: map[string]hashing.Digest{"0|0": []uint8{0x0}},
	}

	inputJSON, _ := json.Marshal(fakeResult)

	serverURL, tearDown := setupServer(inputJSON)
	defer tearDown()
	client := setupClient([]string{serverURL})

	result, err := client.Incremental(start, end)
	assert.NoError(t, err)
	assert.Equal(t, fakeResult, result, "The inputs should match")
}

func TestIncrementalWithServerFailure(t *testing.T) {
	serverURL, tearDown := setupServer(nil)
	defer tearDown()
	client := setupClient([]string{serverURL})

	_, err := client.Incremental(uint64(2), uint64(8))
	assert.Error(t, err)
}

// TODO implement a test to verify proofs using fake hash function

func defaultHandler(input []byte) func(http.ResponseWriter, *http.Request) {
	statusOK := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		out := new(bytes.Buffer)
		_ = json.Compact(out, input)
		_, _ = w.Write(out.Bytes())
	}

	statusInternalServerError := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}

	if input == nil {
		return statusInternalServerError
	} else {
		return statusOK
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
