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
	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

func TestAddBulkAndVerify(t *testing.T) {
	before, after := setupServer(0, "", false, t)
	scenario, let := scope.Scope(t, before, after)
	// log.SetLogger("", log.DEBUG)

	events := []string{
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
		rand.RandomString(10),
	}

	scenario("Add one event and get its membership proof", func() {
		var snapshotBulk []*protocol.Snapshot
		var membershipBulk []*protocol.MembershipResult
		var err error

		client := getClient(t, 0)

		let("Add event", func(t *testing.T) {
			snapshotBulk, err = client.AddBulk(events)
			assert.NoError(t, err)

			for i, snapshot := range snapshotBulk {
				assert.Equal(t, snapshot.EventDigest, hashing.NewSha256Hasher().Do([]byte(events[i])),
					"The snapshot's event doesn't match: expected %s, actual %s", events[i], snapshot.EventDigest)
				assert.False(t, snapshot.Version < 0, "The snapshot's version must be greater or equal to 0")
				assert.False(t, len(snapshot.HyperDigest) == 0, "The snapshot's hyperDigest cannot be empty")
				assert.False(t, len(snapshot.HistoryDigest) == 0, "The snapshot's hyperDigest cannot be empt")
			}
		})

		let("Get membership proof for each inserted event", func(t *testing.T) {
			lastVersion := uint64(len(events) - 1)
			for i, snapshot := range snapshotBulk {

				result, err := client.Membership([]byte(events[i]), snapshot.Version)
				assert.NoError(t, err)

				assert.True(t, result.Exists, "The queried key should be a member")
				assert.Equal(t, result.QueryVersion, snapshot.Version,
					"The query version doest't match the queried one: expected %d, actual %d", snapshot.Version, result.QueryVersion)
				assert.Equal(t, result.ActualVersion, snapshot.Version,
					"The actual version should match the queried one: expected %d, actual %d", snapshot.Version, result.ActualVersion)
				assert.Equal(t, result.CurrentVersion, lastVersion,
					"The current version should match the queried one: expected %d, actual %d", lastVersion, result.CurrentVersion)
				assert.Equal(t, []byte(events[i]), result.Key,
					"The returned event doesn't math the original one: expected %s, actual %s", events[i], result.Key)
				assert.False(t, len(result.KeyDigest) == 0, "The key digest cannot be empty")
				assert.False(t, len(result.Hyper) == 0, "The hyper proof cannot be empty")
				assert.False(t, result.ActualVersion > 0 && len(result.History) == 0,
					"The history proof cannot be empty when version is greater than 0")

				membershipBulk = append(membershipBulk, result)
			}
		})

		let("Verify each membership", func(t *testing.T) {
			for i, result := range membershipBulk {
				snap := snapshotBulk[i]
				assert.True(t, client.DigestVerify(result, snap, hashing.NewSha256Hasher), "result should be valid")
			}
		})
	})

}
