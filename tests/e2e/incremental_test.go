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
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/protocol"
	assert "github.com/stretchr/testify/require"

	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/scope"
)

func TestIncrementalConsistency(t *testing.T) {
	before, after := setup(0, "", t)
	scenario, let := scope.Scope(t, before, after)

	client := getClient(0)

	scenario("Add multiple events and verify consistency between two of them", func() {

		events := make([]string, 10)
		snapshots := make([]*protocol.Snapshot, 10)
		var err error
		var result *protocol.IncrementalResponse

		let("Add ten events", func(t *testing.T) {
			for i := 0; i < 10; i++ {
				events[i] = rand.RandomString(10)
				snapshots[i], err = client.Add(events[i])
				assert.NoError(t, err)
			}
		})

		let("Query for an incremental proof between version 2 and version 8", func(t *testing.T) {
			result, err = client.Incremental(2, 8)
			assert.NoError(t, err)
			assert.Equal(t, uint64(2), result.Start, "The start version should match")
			assert.Equal(t, uint64(8), result.End, "The end version should match")
		})

		let("Verify the proof", func(t *testing.T) {
			assert.True(t, client.VerifyIncremental(result, snapshots[2], snapshots[8], hashing.NewSha256Hasher()), "The proofs should be valid")
		})

	})

}
