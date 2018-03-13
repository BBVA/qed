// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file.

/*
	Package agent implements the command line interface to interact with the
	API rest
*/
package agent

import (
	"context"
	"io/ioutil"
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

	data, err := ioutil.ReadAll(buff)
	check(err)

	glog.Infof("stdin data: %v\n", string(data))

}
