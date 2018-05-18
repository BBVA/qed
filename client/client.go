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
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/balloon"
	"github.com/bbva/qed/balloon/hashing"
)

type HttpClient struct {
	endpoint string
	apiKey   string

	http.Client
}

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

	return bodyBytes, nil

}

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

func (c HttpClient) Membership(key []byte, version uint64) (*apihttp.MembershipProof, error) {

	query, _ := json.Marshal(&apihttp.MembershipQuery{
		key,
		version,
	})

	body, err := c.doReq("POST", "/proofs/membership", query)
	if err != nil {
		return nil, err
	}

	var proof *apihttp.MembershipProof

	json.Unmarshal(body, &proof)

	return proof, nil

}

func (c HttpClient) Verify(proof *apihttp.MembershipProof, snap *apihttp.Snapshot) bool {

	balloonProof := apihttp.ToBalloonProof(c.apiKey, proof, hashing.Sha256Hasher)

	return balloonProof.Verify(&balloon.Commitment{
		snap.HistoryDigest,
		snap.HyperDigest,
		snap.Version,
	}, snap.Event)

}
