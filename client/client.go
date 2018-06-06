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

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/hashing"
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

func (c HttpClient) doReq(method, path string, data []byte) ([]byte, error) {

	req, err := http.NewRequest(method, c.endpoint+path, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.Do(req)
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
func (c HttpClient) Add(event string) (*apihttp.Snapshot, error) {

	data, _ := json.Marshal(&apihttp.Event{[]byte(event)})

	body, err := c.doReq("POST", "/events", data)
	if err != nil {
		return nil, err
	}

	var snapshot apihttp.Snapshot
	json.Unmarshal(body, &snapshot)

	return &snapshot, nil

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

func uint2bytes(i uint64) []byte {
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, i)
	return bytes
}

// Verify will compute the Proof given in Membership and the snapshot from the
// add and returns a proof of existence.

func (c HttpClient) Verify(result *apihttp.MembershipResult, snap *apihttp.Snapshot, hasher hashing.Hasher) bool {

	historyProof, hyperProof := apihttp.ToBalloonProof([]byte(c.apiKey), result, hasher)

	// digest := hasher.Do(snap.Event)
	hyperCorrect := hyperProof.Verify(snap.HyperDigest, result.KeyDigest, uint2bytes(result.QueryVersion))
	if result.Exists {
		if result.QueryVersion <= result.ActualVersion {
			historyCorrect := historyProof.Verify(snap.HistoryDigest, uint2bytes(result.QueryVersion), result.KeyDigest)
			return hyperCorrect && historyCorrect
		}
	}

	return hyperCorrect
}
