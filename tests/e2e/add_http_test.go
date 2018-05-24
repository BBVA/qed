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
	"github.com/bbva/qed/testutils/rand"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddOneEvent(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client := getClient()

	event := rand.RandomString(10)
	snapshot, err := client.Add(event)

	require.NoError(t, err)
	assert.Equal(t, snapshot.Event, []byte(event), "The snapshot's event doesn't match: expected %s, actual %s", event, snapshot.Event)
	assert.False(t, snapshot.Version < 0, "The snapshot's version must be greater or equal to 0")
	assert.False(t, len(snapshot.HyperDigest) == 0, "The snapshot's hyperDigest cannot be empty")
	assert.False(t, len(snapshot.HistoryDigest) == 0, "The snapshot's historyDigest cannot be empty")

}

func TestAddAndQueryOneEvent(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client := getClient()

	event := rand.RandomString(10)
	var snapshot *apihttp.Snapshot

	fmt.Println("add event...")
	snapshot, err := client.Add(event)
	assert.NoError(t, err)

	fmt.Println("query for membership")
	result, err := client.Membership([]byte(event), snapshot.Version)

	require.NoError(t, err)
	assert.True(t, result.IsMember, "The queried key should be a member")
	assert.Equal(t, result.QueryVersion, snapshot.Version, "The query version doest't match the queried one: expected %d, actual %d", snapshot.Version, result.QueryVersion)
	assert.Equal(t, result.ActualVersion, snapshot.Version, "The actual version should match the queried one: expected %d, actual %d", snapshot.Version, result.ActualVersion)
	assert.Equal(t, result.CurrentVersion, snapshot.Version, "The current version should match the queried one: expected %d, actual %d", snapshot.Version, result.CurrentVersion)
	assert.Equal(t, []byte(event), result.Key, "The returned event doesn't math the original one: expected %s, actual %s", event, result.Key)
	assert.False(t, len(result.KeyDigest) == 0, "The key digest cannot be empty")
	assert.False(t, len(result.Proofs.HyperAuditPath) == 0, "The hyper proof cannot be empty")
	assert.False(t, result.ActualVersion > 0 && len(result.Proofs.HistoryAuditPath) == 0, "The history proof cannot be empty when version is greater than 0")

}

func TestAddAndVerifyOneEvent(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client := getClient()

	event := rand.RandomString(10)
	var snapshot *apihttp.Snapshot
	var result *apihttp.MembershipResult

	t.Run("add event", func(t *testing.T) {
		var err error
		snapshot, err = client.Add(event)
		assert.NoError(t, err)
	})

	t.Run("query for membership", func(t *testing.T) {
		var err error
		result, err = client.Membership([]byte(event), snapshot.Version)
		assert.NoError(t, err)
	})

	t.Run("verify proofs", func(t *testing.T) {
		assert.True(t, client.Verify(result, snapshot), "The proofs should be valid")
	})

}

func TestAddTwoEventsAndVerifyFirst(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	client := getClient()

	event1 := rand.RandomString(10)
	event2 := rand.RandomString(10)
	var snapshot1 *apihttp.Snapshot
	var snapshot2 *apihttp.Snapshot
	var result1 *apihttp.MembershipResult

	t.Run("add first event", func(t *testing.T) {
		var err error
		snapshot1, err = client.Add(event1)
		assert.NoError(t, err)
	})

	t.Run("add second event", func(t *testing.T) {
		var err error
		snapshot2, err = client.Add(event2)
		assert.NoError(t, err)
	})

	t.Run("query for membership with first event", func(t *testing.T) {
		var err error
		result1, err = client.Membership([]byte(event1), snapshot1.Version)
		assert.NoError(t, err)
	})

	t.Run("verify proofs", func(t *testing.T) {
		verifyingSnapshot := &apihttp.Snapshot{
			snapshot2.HyperDigest, // note that the hyper digest corresponds with the last one
			snapshot1.HistoryDigest,
			snapshot1.Version,
			snapshot1.Event}
		assert.True(t, client.Verify(result1, verifyingSnapshot), "The proofs should be valid")
	})

}
