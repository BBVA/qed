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
	"fmt"
	"testing"

	"github.com/bbva/qed/api/apihttp"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/publish"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

func TestAddVerify(t *testing.T) {
	before, after := setup(0, "", t)
	scenario, let := scope.Scope(t, before, after)

	client := getClient(0)

	event := rand.RandomString(10)

	scenario("Add one event and get its membership proof", func() {
		var snapshot *publish.Snapshot
		var err error

		let("Add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			assert.NoError(t, err)

			// assert.Equal(t, snapshot.Event, []byte(event), "The snapshot's event doesn't match: expected %s, actual %s", event, snapshot.Snapshot.Event)
			assert.False(t, snapshot.Version < 0, "The snapshot's version must be greater or equal to 0")
			assert.False(t, len(snapshot.HyperDigest) == 0, "The snapshot's hyperDigest cannot be empty")
			assert.False(t, len(snapshot.HistoryDigest) == 0, "The snapshot's hyperDigest cannot be empt")
		})

		let("Get membership proof for first inserted event", func(t *testing.T) {
			result, err := client.Membership([]byte(event), snapshot.Version)
			assert.NoError(t, err)

			assert.True(t, result.Exists, "The queried key should be a member")
			assert.Equal(t, result.QueryVersion, snapshot.Version, "The query version doest't match the queried one: expected %d, actual %d", snapshot.Version, result.QueryVersion)
			assert.Equal(t, result.ActualVersion, snapshot.Version, "The actual version should match the queried one: expected %d, actual %d", snapshot.Version, result.ActualVersion)
			assert.Equal(t, result.CurrentVersion, snapshot.Version, "The current version should match the queried one: expected %d, actual %d", snapshot.Version, result.CurrentVersion)
			assert.Equal(t, []byte(event), result.Key, "The returned event doesn't math the original one: expected %s, actual %s", event, result.Key)
			assert.False(t, len(result.KeyDigest) == 0, "The key digest cannot be empty")
			assert.False(t, len(result.Hyper) == 0, "The hyper proof cannot be empty")
			assert.False(t, result.ActualVersion > 0 && len(result.History) == 0, "The history proof cannot be empty when version is greater than 0")

		})
	})

	scenario("Add two events, verify the first one", func() {
		var result_first, result_last *apihttp.MembershipResult
		var err error
		var first, last *publish.Snapshot

		firstEvent := "Test event 1"
		lastEvent := "Test event 2"
		first, err = client.Add(firstEvent)
		assert.NoError(t, err)
		last, err = client.Add(lastEvent)
		assert.NoError(t, err)

		let("Get membership proof for first inserted event", func(t *testing.T) {
			result_first, err = client.Membership([]byte(firstEvent), first.Version)
			assert.NoError(t, err)
			result_last, err = client.Membership([]byte(lastEvent), last.Version)
			assert.NoError(t, err)
		})

		let("Verify first event", func(t *testing.T) {
			first.HyperDigest = last.HyperDigest
			assert.True(t, client.Verify(result_first, []byte(firstEvent), first, hashing.NewSha256Hasher), "The first proof should be valid")
			assert.True(t, client.Verify(result_last, []byte(lastEvent), last, hashing.NewSha256Hasher), "The last proof should be valid")
		})

	})

	scenario("Add 10 events, verify event with index i", func() {
		var p1, p2 *apihttp.MembershipResult
		var err error
		const size int = 10

		var s [size]*publish.Snapshot

		for i := 0; i < size; i++ {
			s[i], _ = client.Add(fmt.Sprintf("Test Event %d", i))
		}

		i := 3
		j := 6
		k := 9

		let("Get proofs p1, p2 for event with index i in versions j and k", func(t *testing.T) {
			p1, err = client.Membership([]byte(fmt.Sprintf("Test Event %d", i)), s[j].Version)
			assert.NoError(t, err)
			p2, err = client.Membership([]byte(fmt.Sprintf("Test Event %d", i)), s[k].Version)
			assert.NoError(t, err)
		})

		let("Verify both proofs against index i event", func(t *testing.T) {
			snap := &publish.Snapshot{
				s[j].HistoryDigest,
				s[9].HyperDigest,
				s[j].Version,
				// s[i].Event,
			}
			assert.True(t, client.Verify(p1, []byte(fmt.Sprintf("Test Event %d", i)), snap, hashing.NewSha256Hasher), "p1 should be valid")

			snap = &publish.Snapshot{
				s[k].HistoryDigest,
				s[9].HyperDigest,
				s[k].Version,
				// s[i].Event,
			}
			assert.True(t, client.Verify(p2, []byte(fmt.Sprintf("Test Event %d", i)), snap, hashing.NewSha256Hasher), "p2 should be valid")

		})

	})
}
