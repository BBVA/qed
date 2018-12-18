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
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/bbva/qed/hashing"

	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/testutils/scope"
	assert "github.com/stretchr/testify/require"
)

func getSnapshot(version uint64) (*protocol.SignedSnapshot, error) {
	resp, err := http.Get(fmt.Sprintf("%s/snapshot?v=%d", StoreURL, version))
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
	resp, err := http.Get(fmt.Sprintf("%s/alert", StoreURL))
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
	bMonitor, aMonitor := setupMonitor(0, t)
	bPublisher, aPublisher := setupPublisher(0, t)

	scenario, let := scope.Scope(t,
		merge(bServer, bStore, bPublisher, bAuditor, bMonitor),
		merge(aServer, aPublisher, aAuditor, aMonitor, aStore),
	)

	client := getClient(0)
	event := rand.RandomString(10)

	scenario("Add one event and check that it has been published without alerts", func() {
		var snapshot *protocol.Snapshot
		var ss *protocol.SignedSnapshot
		var err error

		let("Add event", func(t *testing.T) {
			snapshot, err = client.Add(event)
			assert.NoError(t, err)
			time.Sleep(2 * time.Second)
		})

		let("Get signed snapshot from snapshot public storage", func(t *testing.T) {
			ss, err = getSnapshot(0)
			assert.NoError(t, err)
			assert.Equal(t, snapshot, ss.Snapshot, "Snapshots must be equal")
		})

		let("Check Auditor do not create any alert", func(t *testing.T) {
			alerts, err := getAlert()
			assert.NoError(t, err)
			assert.False(t, strings.Contains(string(alerts), "Unable to verify snapshot"), "Must not exist alerts")
		})

		let("Check Monitor do not create any alert", func(t *testing.T) {
			alerts, err := getAlert()
			assert.NoError(t, err)
			assert.False(t, strings.Contains(string(alerts), "Unable to verify incremental"), "Must not exist alerts")
		})

	})

	scenario("Add 1st event. Tamper it. Check auditor alerts correctly", func() {
		var err error

		let("Add 1st event", func(t *testing.T) {
			_, err = client.Add(event)
			assert.NoError(t, err)
			time.Sleep(2 * time.Second)
		})

		let("Tamper 1st event", func(t *testing.T) {
			cmd := exec.Command("curl",
				"-sS",
				"-XDELETE",
				"-H", fmt.Sprintf("Api-Key:%s", APIKey),
				"-H", "Content-type: application/json",
				QEDTamperURL,
				"-d", fmt.Sprintf(`{"Digest": "%X"}`, hashing.NewSha256Hasher().Do(hashing.Digest(event))),
			)

			_, err := cmd.CombinedOutput()
			assert.NoError(t, err, "Subprocess must not exit with status 1")
		})

		let("Check Auditor alerts", func(t *testing.T) {
			time.Sleep(1 * time.Second)
			alerts, err := getAlert()
			assert.NoError(t, err)
			assert.True(t, strings.Contains(string(alerts), "Unable to verify snapshot"), "Must exist auditor alerts")
		})

		let("Check Monitor does not create any alert", func(t *testing.T) {
			time.Sleep(1 * time.Second)
			alerts, err := getAlert()
			assert.NoError(t, err)
			assert.False(t, strings.Contains(string(alerts), "Unable to verify incremental"), "Must not exist monitor alert")
		})
	})

	scenario("Add 1st event. Tamper it. Add 2nd event. Check monitor alerts correctly", func() {
		hasher := hashing.NewSha256Hasher()
		tampered := rand.RandomString(10)
		event2 := rand.RandomString(10)

		let("Add 1st event", func(t *testing.T) {
			_, err := client.Add(event)
			assert.NoError(t, err)
			time.Sleep(2 * time.Second)
		})

		let("Tamper 1st event", func(t *testing.T) {
			cmd := exec.Command("curl",
				"-sS",
				"-XPATCH",
				"-H", fmt.Sprintf("Api-Key: %s", APIKey),
				"-H", "Content-type: application/json",
				QEDTamperURL,
				"-d", fmt.Sprintf(`{"Digest": "%X","Value": "%X"}`,
					hasher.Do(hashing.Digest(event)),
					hasher.Do(hashing.Digest(tampered)),
				),
			)

			_, err := cmd.CombinedOutput()

			assert.NoError(t, err, "Subprocess must not exit with status 1")
		})

		let("Add 2nd event", func(t *testing.T) {
			_, err := client.Add(event2)
			assert.NoError(t, err)
			time.Sleep(2 * time.Second)
		})

		let("Check Auditor does not create any alert", func(t *testing.T) {
			alerts, err := getAlert()
			assert.NoError(t, err)
			assert.False(t, strings.Contains(string(alerts), "Unable to verify snapshot"), "Must not exist auditor alerts")
		})

		let("Check Monitor alerts", func(t *testing.T) {
			alerts, err := getAlert()
			assert.NoError(t, err)
			assert.True(t, strings.Contains(string(alerts), "Unable to verify incremental"), "Must exist monitor alert")
		})

	})

	// scenario("Appendix B", func() {
	//
	// 	let("Add 1st event", func(t *testing.T) {
	// 		_, err = client.Add(event)
	// 		assert.NoError(t, err)
	// 		time.Sleep(2 * time.Second)
	// 	})
	// })

}
