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

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/spec"
)

func TestAddVerify(t *testing.T) {
	before, after := newServerSetup(0, false)
	let, report := spec.New()
	defer func() { t.Logf(report()) }()

	event := rand.RandomString(10)
	_, err := before()
	spec.NoError(t, err, "Error starting server")

	let(t, "Add one event and get its membership proof", func(t *testing.T) {
		var snapshot *protocol.Snapshot
		var err error

		client, err := newQedClient(0)
		spec.NoError(t, err, "Error creating qed client")
		defer func() { client.Close() }()

		let(t, "Add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			spec.NoError(t, err, "Error adding event")

			spec.Equal(t, snapshot.EventDigest, hashing.NewSha256Hasher().Do([]byte(event)), "The snapshot's event doesn't match")
			spec.False(t, snapshot.Version < 0, "The snapshot's version must be greater or equal to 0")
			spec.False(t, len(snapshot.HyperDigest) == 0, "The snapshot's hyperDigest cannot be empty")
			spec.False(t, len(snapshot.HistoryDigest) == 0, "The snapshot's hyperDigest cannot be empt")
		})

		let(t, "Get membership proof for first inserted event", func(t *testing.T) {
			proof, err := client.Membership([]byte(event), &snapshot.Version)
			spec.NoError(t, err, "Error getting membership proof")

			spec.True(t, proof.Exists, "The queried key should be a member")
			spec.Equal(t, proof.QueryVersion, snapshot.Version, "The query version doest't match the queried one")
			spec.Equal(t, proof.ActualVersion, snapshot.Version, "The actual version should match the queried one")
			spec.Equal(t, proof.CurrentVersion, snapshot.Version, "The current version should match the queried one")
			spec.False(t, len(proof.KeyDigest) == 0, "The key digest cannot be empty")
			spec.NotNil(t, proof.HyperProof, "The hyper proof cannot be empty")
			spec.False(t, proof.ActualVersion > 0 && proof.HistoryProof == nil, "The history proof cannot be empty when version is greater than 0")
		})
	})
	after()
	_, err = before()
	spec.NoError(t, err, "Error starting server")
	let(t, "Add two events, verify the first one", func(t *testing.T) {
		var proof1, proof2 *balloon.MembershipProof
		var err error
		var snap1, snap2 *protocol.Snapshot

		client, err := newQedClient(0)
		spec.NoError(t, err, "Error creating a new qed client")
		defer func() { client.Close() }()

		snap1, err = client.Add("Test event 1")
		spec.NoError(t, err, "Error adding event 1")
		snap2, err = client.Add("Test event 2")
		spec.NoError(t, err, "Error adding event 2")

		let(t, "Get membership proof for inserted events", func(t *testing.T) {
			proof1, err = client.MembershipDigest(snap1.EventDigest, &snap1.Version)
			spec.NoError(t, err, "Error getting membership digest")
			proof2, err = client.MembershipDigest(snap2.EventDigest, &snap2.Version)
			spec.NoError(t, err, "Error getting membership digest")
		})

		let(t, "Verify events", func(t *testing.T) {
			snap1.HyperDigest = snap2.HyperDigest

			balloonSnap1 := balloon.Snapshot(*snap1)
			balloonSnap2 := balloon.Snapshot(*snap2)
			ok1, _ := client.MembershipVerify(snap1.EventDigest, proof1, &balloonSnap1)
			ok2, _ := client.MembershipVerify(snap2.EventDigest, proof2, &balloonSnap2)
			spec.True(t, ok1, "The first proof should be valid")
			spec.True(t, ok2, "The last proof should be valid")
		})
	})

	after()
	_, err = before()
	spec.NoError(t, err, "Error starting server")
	let(t, "Add 10 events, verify event with index i", func(t *testing.T) {
		var p1, p2 *balloon.MembershipProof
		var err error
		const size int = 10

		var s [size]*protocol.Snapshot

		client, err := newQedClient(0)
		spec.NoError(t, err, "Error creating a new qed client")
		defer func() { client.Close() }()

		for i := 0; i < size; i++ {
			s[i], _ = client.Add(fmt.Sprintf("Test Event %d", i))
		}

		i := 3
		j := 6
		k := 9

		let(t, "Get proofs p1, p2 for event with index i in versions j and k", func(t *testing.T) {
			p1, err = client.MembershipDigest(s[i].EventDigest, &s[j].Version)
			spec.NoError(t, err, "Error getting membership digest")
			p2, err = client.MembershipDigest(s[i].EventDigest, &s[k].Version)
			spec.NoError(t, err, "Error getting membership digest")
		})

		let(t, "Verify both proofs against index i event", func(t *testing.T) {
			snap := &balloon.Snapshot{
				HistoryDigest: s[j].HistoryDigest,
				HyperDigest:   s[9].HyperDigest,
				Version:       s[j].Version,
				EventDigest:   s[i].EventDigest,
			}
			ok, err := client.MembershipVerify(s[i].EventDigest, p1, snap)
			spec.True(t, ok, "p1 should be valid")
			spec.NoError(t, err, "Error not expected")

			snap = &balloon.Snapshot{
				HistoryDigest: s[k].HistoryDigest,
				HyperDigest:   s[9].HyperDigest,
				Version:       s[k].Version,
				EventDigest:   s[i].EventDigest,
			}
			ok, err = client.MembershipVerify(s[i].EventDigest, p2, snap)
			spec.True(t, ok, "p2 should be valid")
			spec.NoError(t, err, "Error not expected")
		})

	})
	after()

}
