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
package e2e

import (
	"fmt"
	"os"
	"testing"

	"github.com/bbva/qed/client"
	"github.com/bbva/qed/server"
	"github.com/bbva/qed/testutils/scope"
)

var apiKey, storageType string
var listenAddr, cacheSize uint64
var getClient func() *client.HttpClient
var incrementalPort uint64

func init() {
	incrementalPort = 1
	listenAddr = 8099
	apiKey = "my-awesome-api-key"
	cacheSize = 50000
	storageType = "badger"
}

func setup() (scope.TestF, scope.TestF) {
	var srv *server.Server
	path := "/var/tmp/balloonE2E"

	port := fmt.Sprintf(":%d", listenAddr+incrementalPort)
	endpoint := fmt.Sprintf("http://127.0.0.1%s", port)
	incrementalPort++

	before := func(t *testing.T) {
		os.RemoveAll(path)
		os.MkdirAll(path, os.FileMode(0755))

		srv = server.NewServer(port, path, apiKey, cacheSize, storageType, false, true)

		go (func() {
			err := srv.Run()
			if err != nil {
				t.Log(err)
			}
		})()
	}

	after := func(t *testing.T) {
		if srv != nil {
			srv.Stop()
		}
	}

	getClient = func() *client.HttpClient {
		return client.NewHttpClient(endpoint, "my-awesome-api-key")
	}

	return before, after

}
