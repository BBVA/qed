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
	"net/http"
	"os"
	"testing"

	"github.com/bbva/qed/client"
	s "github.com/bbva/qed/server"
	"github.com/bbva/qed/testutils/scope"
)

var endpoint, apiKey, storageType, listenAddr string
var cacheLevels uint64

func init() {
	listenAddr = ":8079"
	endpoint = "http://127.0.0.1:8079"
	apiKey = "my-awesome-api-key"
	cacheLevels = 50000
	storageType = "badger"
}

func setup() (scope.TestF, scope.TestF) {
	var server *http.Server
	path := "/var/tmp/balloonE2E"

	before := func(t *testing.T) {
		os.RemoveAll(path)
		os.MkdirAll(path, os.FileMode(0755))

		server = s.NewServer(listenAddr, path, apiKey, cacheLevels, storageType, false, true)

		go (func() {
			err := server.ListenAndServe()
			if err != nil {
				t.Log(err)
			}
		})()

	}

	after := func(t *testing.T) {
		if server != nil {
			server.Close()
		}
	}

	return before, after
}

func getClient() *client.HttpClient {
	return client.NewHttpClient("http://localhost:8079", "my-awesome-api-key")
}