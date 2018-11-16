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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bbva/qed/balloon/visitor"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/sign"
	"github.com/stretchr/testify/assert"
)

var (
	client *HttpClient
	mux    *http.ServeMux
	server *httptest.Server
)

func init() {
	log.SetLogger("client-test", "info")
}

func setup() func() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)
	client = NewHttpClient(server.URL, "my-awesome-api-key")
	return func() {
		server.Close()
	}
}

func TestAddSuccess(t *testing.T) {
	tearDown := setup()
	defer tearDown()

	event := "Hello world!"
	snap := &protocol.Snapshot{
		[]byte("hyper"),
		[]byte("history"),
		0,
		[]byte(event),
	}
	signer := sign.NewEd25519Signer()
	sig, err := signer.Sign([]byte(fmt.Sprintf("%v", snap)))
	fakeSignedSnapshot := &protocol.SignedSnapshot{
		snap,
		sig,
	}

	result, _ := json.Marshal(fakeSignedSnapshot)
	mux.HandleFunc("/events", okHandler(result))

	signedSnapshot, err := client.Add(event)
	assert.NoError(t, err)
	assert.Equal(t, fakeSignedSnapshot, signedSnapshot, "The snapshots should match")

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
		start,
		end,
		visitor.AuditPath{"0|0": []uint8{0x0}},
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
		json.Compact(out, result)
		w.Write(out.Bytes())
	}
}

func serverErrorHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
