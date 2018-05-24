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

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

func TestAddVerify(t *testing.T) {
	before, after := setup()
	scenario, let := scope.Scope(t, before, after)

	client := getClient()

	event := rand.RandomString(10)

	scenario("Add one event and get its membership proof", func() {
		var snapshot *apihttp.Snapshot
		var err error

		let("Add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			assert.NoError(t, err)
			assert.Equal(t, snapshot.Event, []byte(event), "The snapshot's event doesn't match: expected %s, actual %s", event, snapshot.Event)
			assert.False(t, snapshot.Version < 0, "The snapshot's version must be greater or equal to 0")
			assert.False(t, len(snapshot.HyperDigest) == 0, "The snapshot's hyperDigest cannot be empty")
			assert.False(t, len(snapshot.HistoryDigest) == 0, "The snapshot's hyperDigest cannot be empt")
		})

		let("Get membership proof for first inserted event", func(t *testing.T) {

			result, err := client.Membership([]byte(event), snapshot.Version)
			assert.NoError(t, err)
			assert.True(t, result.IsMember, "The queried key should be a member")
			assert.Equal(t, result.QueryVersion, snapshot.Version, "The query version doest't match the queried one: expected %d, actual %d", snapshot.Version, result.QueryVersion)
			assert.Equal(t, result.ActualVersion, snapshot.Version, "The actual version should match the queried one: expected %d, actual %d", snapshot.Version, result.ActualVersion)
			assert.Equal(t, result.CurrentVersion, snapshot.Version, "The current version should match the queried one: expected %d, actual %d", snapshot.Version, result.CurrentVersion)
			assert.Equal(t, []byte(event), result.Key, "The returned event doesn't math the original one: expected %s, actual %s", event, result.Key)
			assert.False(t, len(result.KeyDigest) == 0, "The key digest cannot be empty")
			assert.False(t, len(result.Proofs.HyperAuditPath) == 0, "The hyper proof cannot be empty")
			assert.False(t, result.ActualVersion > 0 && len(result.Proofs.HistoryAuditPath) == 0, "The history proof cannot be empty when version is greater than 0")

		})
	})

	scenario("Add two events, verify the first one", func() {
		var result *apihttp.MembershipResult
		var err error

		first, _ := client.Add("Test event 1")
		last, _ := client.Add("Test event 2")

		let("Get membership proof for first inserted event", func(t *testing.T) {
			result, err = client.Membership(first.Event, first.Version)
			assert.NoError(t, err)
		})

		let("Verify first event", func(t *testing.T) {
			verifyingSnapshot := &apihttp.Snapshot{
				last.HyperDigest, // note that the hyper digest corresponds with the last one
				first.HistoryDigest,
				first.Version,
				first.Event,
			}
			assert.True(t, client.Verify(result, verifyingSnapshot), "The proofs should be valid")

		})

	})

}
