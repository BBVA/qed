// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

/*
	Package agent implements the command line interface to interact with the
	API rest
*/
package agent

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
)

type response map[string]interface{}

type Agent struct {
	httpEndpoint string
	verifier     *balloon.Verifier

	hasher  hashing.Hasher
	storage map[string]*Log
}

type Log struct {
	plain      string
	event      []byte
	commitment *balloon.Commitment
	proof      *balloon.MembershipProof
}

func NewAgent(httpEndpoint string) (*Agent, error) {
	agent := &Agent{
		httpEndpoint,
		balloon.NewDefaultVerifier(),
		hashing.Sha256Hasher,
		make(map[string]*Log),
	}

	// wait some time to load server
	time.Sleep(time.Second)

	return agent, nil

}

func (a *Agent) Add(event string) *balloon.Commitment {
	// TODO: rename message to event also in apiHttp
	data := []byte(strings.Join([]string{`{"message": "`, event, `"}`}, ""))

	req, err := http.NewRequest("POST", a.httpEndpoint+"/events", bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", "this-is-my-api-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	commitment := &balloon.Commitment{}

	json.Unmarshal([]byte(bodyBytes), commitment)

	a.storage[string(a.hasher([]byte(event)))] = &Log{
		plain:      event,
		event:      []byte(event),
		commitment: commitment,
	}

	return commitment

}

func (a *Agent) MembershipProof(event []byte, version uint) *balloon.MembershipProof {
	req, err := http.NewRequest("GET", a.httpEndpoint+"/proofs/membership", nil)
	if err != nil {
		panic(err)
	}

	q := req.URL.Query()
	q.Set("event", string(event))
	q.Set("version", string(version))
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", "this-is-my-api-key")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	proof := &balloon.MembershipProof{}

	json.Unmarshal([]byte(bodyBytes), proof)

	a.storage[string(a.hasher(event))].proof = proof

	return proof
}

func (a *Agent) Verify(proof *balloon.MembershipProof, commitment *balloon.Commitment, event []byte) bool {
	result, err := a.verifier.Verify(proof, commitment, event)
	if err != nil {
		// TODO: log error internally
		return false
	}

	return result
}
