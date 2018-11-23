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
package publisher

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/valyala/fasthttp"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

func TestProcess(t *testing.T) {
	// Fake server
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer fakeServer.Close()

	// Batch
	var batch protocol.BatchSnapshots
	signedSnapshot := protocol.SignedSnapshot{
		Snapshot: &protocol.Snapshot{
			HistoryDigest: []byte{0x0},
			HyperDigest:   []byte{0x0},
			Version:       0,
		},
		Signature: []byte{0x0},
	}
	batch.Snapshots = append(batch.Snapshots, &signedSnapshot)
	batch.TTL = 0

	// Publisher
	conf := NewConfig(&fasthttp.Client{}, []string{fakeServer.URL})
	p := NewPublisher(conf)
	p.Process(&batch)
}

func BenchmarkPublisher(b *testing.B) {
	log.SetLogger("BenchmarkPublisher", log.SILENT)

	// Fake server
	fakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer fakeServer.Close()

	// Batch
	var batch protocol.BatchSnapshots
	signedSnapshot := protocol.SignedSnapshot{
		Snapshot: &protocol.Snapshot{
			HistoryDigest: []byte{0x0},
			HyperDigest:   []byte{0x0},
			Version:       0,
		},
		Signature: []byte{0x0},
	}
	batch.Snapshots = append(batch.Snapshots, &signedSnapshot)
	batch.TTL = 0

	// Publisher
	conf := NewConfig(&fasthttp.Client{}, []string{fakeServer.URL})
	p := NewPublisher(conf)

	// Bench
	b.ResetTimer()
	b.N = 100000
	for i := 0; i < b.N; i++ {
		p.Process(&batch)
	}
}
