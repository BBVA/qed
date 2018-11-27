/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, n.A.
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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
)

type Processor interface {
	Process(*protocol.BatchSnapshots)
}

type FakeProcessor struct {
}

func (d FakeProcessor) Process(b *protocol.BatchSnapshots) {
}

type DummyProcessor struct {
}

func (d DummyProcessor) Process(b *protocol.BatchSnapshots) {
	for i := 0; i < len(b.Snapshots); i++ {
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:8888/?nodeType=auditor&id=%d", b.Snapshots[0].Snapshot.Version))
		if err != nil || res == nil {
			log.Debugf("Error contacting service with error %v", err)
		}
		// to reuse connections we need to do this
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()

		// time.Sleep(1 * time.Second)
	}

	log.Debugf("process(): Processed %v elements of batch id %v", len(b.Snapshots), b.Snapshots[0].Snapshot.Version)
}
