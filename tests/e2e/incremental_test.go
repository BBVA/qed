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
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/protocol"

	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/spec"
)

func TestIncrementalConsistency(t *testing.T) {
	before, after := newServerSetup(0, false)
	let, report := spec.New()
	defer func() {
		after()
		t.Logf(report())
	}()
	err := before()
	spec.NoError(t, err, "Error starting server")

	let(t, "Add multiple events and verify consistency between two of them", func(t *testing.T) {
		client, err := newQedClient(0)
		spec.NoError(t, err, "Error creating a new qed client")
		defer func(){ client.Close() }()
		events := make([]string, 10)
		snapshots := make([]*protocol.Snapshot, 10)
		var result *protocol.IncrementalResponse

		let(t, "Add ten events", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				events[i] = rand.RandomString(10)
				snapshots[i], err = client.Add(events[i])
				spec.NoError(t, err, "Error adding event")
			}
		})

		let(t, "Query for an incremental proof between version 2 and version 8", func(t *testing.T) {
			result, err = client.Incremental(2, 8)
			spec.NoError(t, err, "error getting incremental proof")
			spec.Equal(t, uint64(2), result.Start, "The start version should match")
			spec.Equal(t, uint64(8), result.End, "The end version should match")
		})

		let(t, "Verify the proof", func(t *testing.T) {
			spec.True(t, client.VerifyIncremental(result, snapshots[2], snapshots[8], hashing.NewSha256Hasher()), "The proofs should be valid")
		})

	})

}
