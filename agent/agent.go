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
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/glog"
)

type Agent struct {
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func Run(ctx context.Context) (*Agent, error) {
	agent := new(Agent)
	// mock some time

	time.Sleep(time.Second * 2)

	return agent, nil

}

func (a *Agent) Echo(buff *os.File) {

	// data, err := ioutil.ReadAll(buff)
	// check(err)

	data := []byte(`{"message": "this is a sample event"}`)

	req, err := http.NewRequest("POST", "http://localhost:8080/events", bytes.NewBuffer(data))
	check(err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", "this-is-my-api-key")

	resp, err := http.DefaultClient.Do(req)
	check(err)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	check(err)

	glog.Infof("stdin data: %v\n", string(body))

}
