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
	"net/http"
	"strings"
	"time"
)

type Agent struct{}

func Run(ctx context.Context) (*Agent, error) {
	agent := new(Agent)

	// wait some time to load server
	time.Sleep(time.Second)

	return agent, nil

}

func (a *Agent) Add(message string) {
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

}
