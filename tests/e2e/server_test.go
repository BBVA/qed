/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"testing"
	"time"

	"github.com/bbva/qed/testutils/scenario"
)

func TestStart(t *testing.T) {
	before, after := prepare_new_server(0, "", false)
	let, report := scenario.New()
	defer func() {
		after()
		t.Logf(report())
	}()
	before()

	let(t, "Test availability of qed server", func(t *testing.T) {
		let(t, "Query info endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 2*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8800/info", APIKey, nil)
				return err
			})
			scenario.NoError(t, err, "Subprocess must not exit with non-zero status")
			scenario.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

		let(t, "Query to unexpected context", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8800/xD", APIKey, nil)
				return err
			})
			scenario.NoError(t, err, "Error getting response from server")
			scenario.Equal(t, resp.StatusCode, http.StatusNotFound, "Server should respond with http status code 404")

		})
	})

	let(t, "Test availability of metrics server", func(t *testing.T) {
		let(t, "Query metrics endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8600/metrics", APIKey, nil)
				return err
			})
			scenario.NoError(t, err, "Subprocess must not exit with non-zero status")
			scenario.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

	})

}
