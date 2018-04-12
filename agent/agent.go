// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file.

/*
	Package agent implements the command line interface to interact with the
	API rest
*/
package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	apiHttp "verifiabledata/api/apihttp"
)

type Agent struct{}

func Run(ctx context.Context) (*Agent, error) {
	agent := new(Agent)

	// wait some time to load server
	time.Sleep(time.Second)

	return agent, nil

}

func (a *Agent) Add(message string) *apiHttp.InsertResponse {
	data := []byte(strings.Join([]string{`{"message": "`, message, `"}`}, ""))

	req, err := http.NewRequest("POST", "http://localhost:8080/events", bytes.NewBuffer(data))
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	insert := &apiHttp.InsertResponse{}

	json.Unmarshal([]byte(bodyBytes), &insert)

	return insert
}

func (a *Agent) Fetch(message string) *apiHttp.FetchResponse {
	data := []byte(strings.Join([]string{`{"message": "`, message, `"}`}, ""))

	// Create a simple request to out fetch endpoint
	req, err := http.NewRequest("GET", "http://localhost:8080/fetch", bytes.NewBuffer(data))
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

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	fetch := &apiHttp.FetchResponse{}

	json.Unmarshal([]byte(bodyBytes), &fetch)

	return fetch
}

func (a *Agent) Verify(message string) {
	return
}
