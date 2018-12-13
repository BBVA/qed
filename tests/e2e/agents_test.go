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
	"strings"
	"testing"
	"time"

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

func getAlert() ([]byte, error) {
	resp, err := http.Get(fmt.Sprintf("%s/alert", StoreUrl))
	if err != nil {
		return []byte{}, fmt.Errorf("Error getting alert from alertStore: %v", err)
	}
	defer resp.Body.Close()
	alerts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Error parsing alert from alertStore: %v", err)
	}
	return alerts, nil
}

func TestAgents(t *testing.T) {
	bStore, aStore := setupStore(t)
	bServer, aServer := setupServer(0, "", t)
	bAuditor, aAuditor := setupAuditor(0, t)
	// bMonitor, aMonitor := setupMonitor(0, t)
	bPublisher, aPublisher := setupPublisher(0, t)

	scenario, let := scope.Scope(t,
		merge(bStore, bServer, bAuditor /* bMonitor*/, bPublisher),
		merge(aAuditor /*,aMonitor*/, aPublisher, aServer, aStore),
	)

	client := getClient(0)
	event := rand.RandomString(10)

	scenario("Add one event and check that it has been published", func() {
		var snapshot *protocol.Snapshot
		var ss *protocol.SignedSnapshot
		var err error

		let("Add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			assert.NoError(t, err)
		})

		let("Get signed snapshot from snapshot public storage", func(t *testing.T) {
			time.Sleep(1 * time.Second)
			ss, err = getSnapshot(0)
			assert.NoError(t, err)
			assert.Equal(t, snapshot, ss.Snapshot, "Snapshots must be equal")
		})

		let("Check Auditor do not create any alert", func(t *testing.T) {
			time.Sleep(1 * time.Second)
			alerts, err := getAlert()
			fmt.Println(string(alerts))
			assert.NoError(t, err)
			assert.False(t, strings.Contains(string(alerts), "Unable to verify"), "Must not exist alerts")
		})
	})

}
