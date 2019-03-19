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

	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

func TestStart(t *testing.T) {
	bServer, aServer := setupServer(0, "", false, t)

	scenario, let := scope.Scope(t,
		merge(bServer),
		merge(aServer, delay(2*time.Second)),
	)

	scenario("Test availability of qed server", func() {
		let("Query metrics endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8800/info", APIKey, nil)
				return err
			})
			assert.NoError(t, err, "Subprocess must not exit with non-zero status")
			assert.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

		let("Query to unexpected context", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8800/xD", APIKey, nil)
				return err
			})
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, http.StatusNotFound, "Server should respond with http status code 404")

		})
	})

	scenario("Test availability of metrics server", func() {
		let("Query metrics endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8600/metrics", APIKey, nil)
				return err
			})
			assert.NoError(t, err, "Subprocess must not exit with non-zero status")
			assert.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

	})

}
