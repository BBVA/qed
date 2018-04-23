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
)

type response map[string]interface{}

type Agent struct {
	httpEndpoint string
}

func NewAgent(httpEndpoint string) (*Agent, error) {
	agent := &Agent{
		httpEndpoint,
	}

	// wait some time to load server
	time.Sleep(time.Second)

	return agent, nil

}

func (a *Agent) Add(message string) *balloon.Commitment {
	data := []byte(strings.Join([]string{`{"message": "`, message, `"}`}, ""))

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

	return commitment

}

func (a *Agent) MembershipProof(commitment *balloon.Commitment) *balloon.MembershipProof {

	data, err := json.Marshal(commitment)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", a.httpEndpoint+"/proofs/membership", bytes.NewBuffer(data))
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

	proof := &balloon.MembershipProof{}

	json.Unmarshal([]byte(bodyBytes), proof)

	return proof
}

func (a *Agent) Verify(proof *balloon.MembershipProof) bool {

	result, err := balloon.Verify(proof)
	if err != nil {
		// TODO: log error internally
		return false
	}

	return result
}
