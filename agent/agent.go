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

func (a *Agent) Add(message string) response {
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

	insert := make(response)

	json.Unmarshal([]byte(bodyBytes), &insert)

	return insert
}

func (a *Agent) Verify(message string) {
	return
}
