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

package gossip

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bbva/qed/log"
	"github.com/stretchr/testify/require"
)

func TestDefaultAlert(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		called = true
		w.WriteHeader(http.StatusNoContent)
	}))

	defer server.Close()
	conf := DefaultSimpleNotifierConfig()
	conf.Endpoint = append(conf.Endpoint, server.URL)
	notificator := NewSimpleNotifierFromConfig(conf, log.L())

	notificator.Start()
	defer notificator.Stop()

	_ = notificator.Alert("test alert")
	time.Sleep(1 * time.Second)

	require.True(t, called, "Server must be called from alerter")
}
