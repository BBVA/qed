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

// Package client implements the command line interface to interact with
// the REST API
package client

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/publish"
)

// HttpClient ist the stuct that has the required information for the cli.
type HttpClient struct {
	endpoint string
	apiKey   string

	http.Client
}

// NewHttpClient will return a new instance of HttpClient.
func NewHttpClient(endpoint, apiKey string) *HttpClient {

	return &HttpClient{
		endpoint,
		apiKey,
		*http.DefaultClient,
	}

}

func (c HttpClient) exponentialBackoff(req *http.Request) (*http.Response, error) {

	var retries uint

	for {
		resp, err := c.Do(req)
		if err != nil {
			if retries == 5 {
				return nil, err
			}
			retries = retries + 1
			delay := time.Duration(10 << retries * time.Millisecond)
			time.Sleep(delay)
			continue
		}
		return resp, err
	}

}

func (c HttpClient) doReq(method, path string, data []byte) ([]byte, error) {

	req, err := http.NewRequest(method, c.endpoint+path, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.exponentialBackoff(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("Unexpected server error")
	}

	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return nil, fmt.Errorf("Invalid request")
	}

	return bodyBytes, nil

}

// Add will do a request to the server with a post data to store a new event.
func (c HttpClient) Add(event string) error {

	data, _ := json.Marshal(&apihttp.Event{[]byte(event)})

	_, err := c.doReq("POST", "/events", data)
	if err != nil {
		return err
	}

	return nil
}

// Membership will ask for a Proof to the server.
func (c HttpClient) Membership(key []byte, version uint64) (*apihttp.MembershipResult, error) {

	query, _ := json.Marshal(&apihttp.MembershipQuery{
		key,
		version,
	})

	body, err := c.doReq("POST", "/proofs/membership", query)
	if err != nil {
		return nil, err
	}

	var proof *apihttp.MembershipResult
	json.Unmarshal(body, &proof)

	return proof, nil

}

// Incremental will ask for an IncrementalProof to the server.
func (c HttpClient) Incremental(start, end uint64) (*apihttp.IncrementalResponse, error) {

	query, _ := json.Marshal(&apihttp.IncrementalRequest{
		start,
		end,
	})

	body, err := c.doReq("POST", "/proofs/incremental", query)
	if err != nil {
		return nil, err
	}

	var response *apihttp.IncrementalResponse
	json.Unmarshal(body, &response)

	return response, nil
}

func uint2bytes(i uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, i)
	return bytes
}

// Verify will compute the Proof given in Membership and the snapshot from the
// add and returns a proof of existence.

func (c HttpClient) Verify(result *apihttp.MembershipResult, event []byte, snap *publish.Snapshot, hasherF func() hashing.Hasher) bool {

	proof := apihttp.ToBalloonProof([]byte(c.apiKey), result, hasherF)

	return proof.Verify(event, &balloon.Commitment{
		snap.HistoryDigest,
		snap.HyperDigest,
		snap.Version,
	})

}

func (c HttpClient) VerifyIncremental(result *apihttp.IncrementalResponse, startSnapshot, endSnapshot *publish.SignedSnapshot, hasher hashing.Hasher) bool {

	proof := apihttp.ToIncrementalProof(result, hasher)

	startCommitment := &balloon.Commitment{
		startSnapshot.Snapshot.HistoryDigest,
		startSnapshot.Snapshot.HyperDigest,
		startSnapshot.Snapshot.Version,
	}
	endCommitment := &balloon.Commitment{
		endSnapshot.Snapshot.HistoryDigest,
		endSnapshot.Snapshot.HyperDigest,
		endSnapshot.Snapshot.Version,
	}

	return proof.Verify(startCommitment, endCommitment)

}
