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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/testutils/spec"
)

// This test is used to profile server starting time
// go test -v -cpuprofile=cpu2.out ./... -run ProfilingStart
func TestProfilingStart(t *testing.T) {
	before, after := newServerSetup(0, false)
	defer after()
	err := before()
	spec.NoError(t, err, "Error starting server")
	<-time.After(10 * time.Second)
}

func TestStart(t *testing.T) {
	before, after := newServerSetup(0, false)
	let, report := spec.New()
	defer func() {
		after()
		t.Logf(report())
	}()
	err := before()
	spec.NoError(t, err, "Error starting server")
	let(t, "Test availability of qed server", func(t *testing.T) {
		let(t, "Query info endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 2*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8800/info", "APIKey", nil)
				return err
			})
			spec.NoError(t, err, "Subprocess must not exit with non-zero status")
			spec.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

		let(t, "Query to unexpected context", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8800/xD", "APIKey", nil)
				return err
			})
			spec.NoError(t, err, "Error getting response from server")
			spec.Equal(t, resp.StatusCode, http.StatusNotFound, "Server should respond with http status code 404")

		})
	})

	let(t, "Test availability of metrics server", func(t *testing.T) {
		let(t, "Query metrics endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8600/metrics", "APIKey", nil)
				return err
			})
			spec.NoError(t, err, "Subprocess must not exit with non-zero status")
			spec.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

	})

}
func TestStartTls(t *testing.T) {
	before, after := newServerSetup(0, true)
	let, report := spec.New()
	defer func() {
		after()
		t.Logf(report())
	}()
	err := before()
	spec.NoError(t, err, "Error starting server")
	let(t, "Test availability of qed server", func(t *testing.T) {
		let(t, "Query info endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 2*time.Second, func() error {
				resp, err = doReq("GET", "https://localhost:8800/info", "APIKey", nil)
				return err
			})
			spec.NoError(t, err, "Subprocess must not exit with non-zero status")
			spec.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

		let(t, "Query to unexpected context", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "https://localhost:8800/xD", "APIKey", nil)
				return err
			})
			spec.NoError(t, err, "Error getting response from server")
			spec.Equal(t, resp.StatusCode, http.StatusNotFound, "Server should respond with http status code 404")

		})
	})

	let(t, "Test availability of metrics server", func(t *testing.T) {
		let(t, "Query metrics endpoint", func(t *testing.T) {
			var resp *http.Response
			var err error
			retry(3, 1*time.Second, func() error {
				resp, err = doReq("GET", "http://localhost:8600/metrics", "APIKey", nil)
				return err
			})
			spec.NoError(t, err, "Subprocess must not exit with non-zero status")
			spec.Equal(t, resp.StatusCode, http.StatusOK, "Server should respond with http status code 200")
		})

	})

}
func TestStartCluster(t *testing.T) {
	b0, a0 := newServerSetup(0, false)
	b1, a1 := newServerSetup(1, false)
	b2, a2 := newServerSetup(2, false)
	let, report := spec.New()
	defer func() {
		a0()
		a1()
		a2()
		t.Logf(report())
	}()
	log.SetLogger("e2e", log.DEBUG)

	let(t, "Start three servers", func(t *testing.T) {
		err := b0()
		spec.NoError(t, err, "Error starting node 1")
		err = b1()
		spec.NoError(t, err, "Error starting node 2")
		err = b2()
		spec.NoError(t, err, "Error starting node 3")
	})

	let(t, "Check the cluster topology", func(t *testing.T) {
		var resp *http.Response
		var mainErr error
		retry(3, 2*time.Second, func() error {
			var subErr error
			retry(3, 2*time.Second, func() error {
				resp, subErr = doReq("GET", "http://localhost:8800/info/shards", "APIKey", nil)
				return subErr
			})
			if subErr != nil {
				return subErr
			}
			if resp == nil {
				mainErr = fmt.Errorf("nil response")
				return mainErr
			}

			buff, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			m := make(map[string]interface{})
			err = json.Unmarshal(buff, &m)
			if err != nil {
				mainErr = fmt.Errorf("error decoding info: %v", err)
				return mainErr
			}
			shards := len(m["shards"].(map[string]interface{}))
			if shards != 3 {
				mainErr = fmt.Errorf("not enought shards: %v", shards)
				return mainErr
			}
			return nil
		})
		spec.NoError(t, mainErr, "There should be no error")
	})
}
