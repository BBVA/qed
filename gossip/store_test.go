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

	"github.com/stretchr/testify/require"
)

func storeHandler(called *bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.Header().Set("Allow", "POST")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		*called = true
		w.WriteHeader(http.StatusNoContent)
	}
}

func TestDefaultStore(t *testing.T) {
	var called bool

	server := httptest.NewServer(storeHandler(&called))
	defer server.Close()
	notificator := NewDefaultNotifier([]string{server.URL}, 0, 0, 0)

	notificator.Start()
	notificator.Alert("test alert")
	time.Sleep(1 * time.Second)
	notificator.Stop()
	require.True(t, called, "Server must be called from alerter")
}
