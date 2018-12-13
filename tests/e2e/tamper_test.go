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
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	// "github.com/bbva/qed/hashing"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

func getSnapshot(version uint64) (*protocol.SignedSnapshot, error) {
	resp, err := http.Get(fmt.Sprintf("%s/snapshot?v=%d", StoreUrl, version))
	if err != nil {
		return nil, fmt.Errorf("Error getting snapshot from the store: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Error getting snapshot from the store. Status: %d", resp.StatusCode)
	}

	buf, _ := ioutil.ReadAll(resp.Body)
	s := &protocol.SignedSnapshot{}
	err = s.Decode(buf)
	if err != nil {
		return nil, fmt.Errorf("Error decoding signed snapshot %d codec", version)
	}
	return s, nil
}

func TestTamper(t *testing.T) {
	bStore, aStore := setupStore(t)
	bServer, aServer := setupServer(0, "", t)
	// bAuditor, aAuditor := setupAuditor(0, t)
	// bMonitor, aMonitor := setupMonitor(0, t)
	bPublisher, aPublisher := setupPublisher(0, t)

	scenario, let := scope.Scope(t,
		merge(bStore, bServer /*bAuditor, bMonitor*/, bPublisher),
		merge( /*aAuditor, aMonitor,*/ aPublisher, aServer, aStore),
	)

	client := getClient(0)

	event := rand.RandomString(10)

	// scenario("Add one event and get its membership proof", func() {
	// 	var snapshot *protocol.Snapshot
	// 	var err error

	// 	let("Add event", func(t *testing.T) {
	// 		snapshot, err = client.Add(event)
	// 		assert.NoError(t, err)

	// 		assert.Equal(t, snapshot.EventDigest, hashing.NewSha256Hasher().Do([]byte(event)),
	// 			"The snapshot's event doesn't match: expected %s, actual %s", event, snapshot.EventDigest)
	// 		assert.False(t, snapshot.Version < 0, "The snapshot's version must be greater or equal to 0")
	// 		assert.False(t, len(snapshot.HyperDigest) == 0, "The snapshot's hyperDigest cannot be empty")
	// 		assert.False(t, len(snapshot.HistoryDigest) == 0, "The snapshot's hyperDigest cannot be empt")
	// 	})

	// 	let("Get membership proof for first inserted event", func(t *testing.T) {
	// 		result, err := client.Membership([]byte(event), snapshot.Version)
	// 		assert.NoError(t, err)

	// 		assert.True(t, result.Exists, "The queried key should be a member")
	// 		assert.Equal(t, result.QueryVersion, snapshot.Version,
	// 			"The query version doest't match the queried one: expected %d, actual %d", snapshot.Version, result.QueryVersion)
	// 		assert.Equal(t, result.ActualVersion, snapshot.Version,
	// 			"The actual version should match the queried one: expected %d, actual %d", snapshot.Version, result.ActualVersion)
	// 		assert.Equal(t, result.CurrentVersion, snapshot.Version,
	// 			"The current version should match the queried one: expected %d, actual %d", snapshot.Version, result.CurrentVersion)
	// 		assert.Equal(t, []byte(event), result.Key,
	// 			"The returned event doesn't math the original one: expected %s, actual %s", event, result.Key)
	// 		assert.False(t, len(result.KeyDigest) == 0, "The key digest cannot be empty")
	// 		assert.False(t, len(result.Hyper) == 0, "The hyper proof cannot be empty")
	// 		assert.False(t, result.ActualVersion > 0 && len(result.History) == 0,
	// 			"The history proof cannot be empty when version is greater than 0")

	// 	})
	// })

	scenario("Add one event and check that it has been published", func() {
		var snapshot *protocol.Snapshot
		var err error

		let("Add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			assert.NoError(t, err)
		})

		let("Get signed snapshot from snapshot public storage", func(t *testing.T) {
			time.Sleep(2 * time.Second)
			ss, err := getSnapshot(0)
			if err != nil {
				fmt.Println("Error: ", err)
			}
			assert.Equal(t, snapshot, ss.Snapshot)
		})
	})

}
