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

package apihttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/sign"

	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/balloon/proof"
	assert "github.com/stretchr/testify/require"
)

type fakeBalloon struct {
	addch   chan *balloon.Commitment
	stopch  chan bool
	proofch chan *balloon.MembershipProof
	incch   chan *balloon.IncrementalProof
}

func (b fakeBalloon) Add(event []byte) chan *balloon.Commitment {
	return b.addch
}

func (b fakeBalloon) Close() chan bool {
	return b.stopch
}

func (b fakeBalloon) GenMembershipProof(event []byte, version uint64) chan *balloon.MembershipProof {
	return b.proofch
}

func (b fakeBalloon) GenIncrementalProof(start, end uint64) chan *balloon.IncrementalProof {
	return b.incch
}

func newAddOpFakeBalloon(addch chan *balloon.Commitment) fakeBalloon {
	closech := make(chan bool)
	proofch := make(chan *balloon.MembershipProof)
	incch := make(chan *balloon.IncrementalProof)

	return fakeBalloon{
		addch,
		closech,
		proofch,
		incch,
	}
}

func newMembershipOpFakeBalloon(proofch chan *balloon.MembershipProof) fakeBalloon {
	addch := make(chan *balloon.Commitment)
	incch := make(chan *balloon.IncrementalProof)
	closech := make(chan bool)

	return fakeBalloon{
		addch,
		closech,
		proofch,
		incch,
	}
}

func newIncrementalOpFakeBalloon(incch chan *balloon.IncrementalProof) fakeBalloon {
	addch := make(chan *balloon.Commitment)
	closech := make(chan bool)
	proofch := make(chan *balloon.MembershipProof)

	return fakeBalloon{
		addch,
		closech,
		proofch,
		incch,
	}
}

func TestHealthCheckHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"version":0,"status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestAdd(t *testing.T) {
	// Create a request to pass to our handler. We pass a message as a data.
	// If it's nil it will fail.
	data, _ := json.Marshal(&Event{[]byte("this is a sample event")})

	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(data))
	if len(data) == 0 {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	addch := make(chan *balloon.Commitment)
	signer := sign.NewSigner()
	handler := Add(newAddOpFakeBalloon(addch), signer)

	go func() {
		addch <- &balloon.Commitment{
			[]byte{0x0},
			[]byte{0x1},
			0,
		}
	}()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Check the body response
	signedSnapshot := &SignedSnapshot{}

	json.Unmarshal([]byte(rr.Body.String()), signedSnapshot)

	if !bytes.Equal(signedSnapshot.Snapshot.HyperDigest, []byte{0x1}) {
		t.Errorf("HyperDigest is not consistent: %s", signedSnapshot.Snapshot.HyperDigest)
	}

	if !bytes.Equal(signedSnapshot.Snapshot.HistoryDigest, []byte{0x0}) {
		t.Errorf("HistoryDigest is not consistent %s", signedSnapshot.Snapshot.HistoryDigest)
	}

	if signedSnapshot.Snapshot.Version != 0 {
		t.Errorf("Version is not consistent")
	}

	if !bytes.Equal(signedSnapshot.Snapshot.Event, []byte("this is a sample event")) {
		t.Errorf("Event is not consistent ")
	}

	signature, err := signer.Sign([]byte(fmt.Sprintf("%v", signedSnapshot.Snapshot)))

	if !bytes.Equal(signedSnapshot.Signature, signature) {
		t.Errorf("Signature is not consistent")
	}

}

func TestMembership(t *testing.T) {
	var version uint64 = 1
	key := []byte("this is a sample event")
	query, _ := json.Marshal(MembershipQuery{
		key,
		version,
	})

	req, err := http.NewRequest("POST", "/proof/membership", bytes.NewBuffer(query))
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	p := make(chan *balloon.MembershipProof)
	handler := Membership(newMembershipOpFakeBalloon(p))
	expectedResult := &MembershipResult{Exists: true, Hyper: map[string][]uint8(nil), History: map[string][]uint8(nil), CurrentVersion: 0x1, QueryVersion: 0x1, ActualVersion: 0x2, KeyDigest: []uint8{0x0}, Key: []uint8{0x74, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x20, 0x65, 0x76, 0x65, 0x6e, 0x74}}
	go func() {
		p <- balloon.NewMembershipProof(
			true,
			proof.NewProof(nil, nil, nil),
			proof.NewProof(nil, nil, nil),
			version,
			version,
			version+1,
			[]byte{0x0},
			new(hashing.Sha256Hasher),
		)
	}()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the body response
	actualResult := new(MembershipResult)
	json.Unmarshal([]byte(rr.Body.String()), actualResult)

	assert.Equal(t, expectedResult, actualResult, "Incorrect proof")

}

func TestIncremental(t *testing.T) {
	start := uint64(2)
	end := uint64(8)
	query, _ := json.Marshal(IncrementalRequest{
		start,
		end,
	})

	req, err := http.NewRequest("POST", "/proof/incremental", bytes.NewBuffer(query))
	assert.NoError(t, err, "Error querying for incremental proof")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	incch := make(chan *balloon.IncrementalProof)
	handler := Incremental(newIncrementalOpFakeBalloon(incch))
	expectedResult := &IncrementalResponse{
		start,
		end,
		map[string][]byte{"0|0": []uint8{0x0}},
	}
	go func() {
		incch <- balloon.NewIncrementalProof(
			start, 
			end, 
			proof.AuditPath{"0|0": []uint8{0x0}},
			new(hashing.Sha256Hasher),
		)
	}()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	status := rr.Code
	assert.Equalf(t, http.StatusOK, status, "handler returned wrong status code: got %v want %v", status, http.StatusOK)

	// Check the body response
	actualResult := new(IncrementalResponse)
	json.Unmarshal([]byte(rr.Body.String()), actualResult)

	assert.Equal(t, expectedResult, actualResult, "Incorrect proof")
}

func TestAuthHandlerMiddleware(t *testing.T) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set Api-Key header
	req.Header.Set("Api-Key", "this-is-my-api-key")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := AuthHandlerMiddleware(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func BenchmarkNoAuth(b *testing.B) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		b.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Define our http client
	client := &http.Client{}

	for i := 0; i < b.N; i++ {
		client.Do(req)
	}
}

func BenchmarkAuth(b *testing.B) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		b.Fatal(err)
	}

	// Set Api-Key header
	req.Header.Set("Api-Key", "this-is-my-api-key")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := AuthHandlerMiddleware(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Define our http client
	client := &http.Client{}

	for i := 0; i < b.N; i++ {
		client.Do(req)
	}
}
