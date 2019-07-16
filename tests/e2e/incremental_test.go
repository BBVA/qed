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
	"fmt"
	"testing"

	"github.com/bbva/qed/balloon"
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
	_, err := before()
	spec.NoError(t, err, "Error starting server")

	let(t, "Add multiple events and verify consistency between two of them", func(t *testing.T) {
		client, err := newQedClient(0)
		spec.NoError(t, err, "Error creating a new qed client")
		defer func() { client.Close() }()
		events := make([]string, 10)
		snapshots := make([]*protocol.Snapshot, 10)
		var proof *balloon.IncrementalProof

		let(t, "Add ten events", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				events[i] = rand.RandomString(10)
				snapshots[i], err = client.Add(events[i])
				spec.NoError(t, err, "Error adding event")
			}
		})

		let(t, "Query for an incremental proof between version 2 and version 8", func(t *testing.T) {
			proof, err = client.Incremental(2, 8)
			spec.NoError(t, err, "error getting incremental proof")
			spec.Equal(t, uint64(2), proof.Start, "The start version should match")
			spec.Equal(t, uint64(8), proof.End, "The end version should match")
		})

		let(t, "Verify the proof", func(t *testing.T) {
			balloonStartSnapshot := balloon.Snapshot(*snapshots[2])
			balloonEndSnapshot := balloon.Snapshot(*snapshots[8])
			ok, err := client.IncrementalVerify(proof, &balloonStartSnapshot, &balloonEndSnapshot)
			spec.True(t, ok, "The proofs should be valid")
			spec.NoError(t, err, fmt.Sprintf("Unexpected error: %s", err))
		})

	})

}
