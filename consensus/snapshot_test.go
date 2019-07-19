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

package consensus

import (
	"testing"
	"time"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/stretchr/testify/require"
)

func TestSnapshot(t *testing.T) {

	log.SetLogger("TestSnapshot", log.SILENT)

	// start only one seed
	node, clean, err := newSeed(1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node.Close(true))
		clean()
	}()

	// add some events
	hasher := hashing.NewSha256Hasher()
	events := append([]hashing.Digest{},
		hasher.Do([]byte("The year’s at the spring,")),
	)
	cmd := newCommand(addEventCommandType)
	cmd.encode(events)

	// apply to WAL
	resp := node.Apply(newLog(0, 0, cmd.data))
	require.NoError(t, resp.(*fsmResponse).err)

	// take a snapshot to inspect its values
	snap, err := node.Snapshot()
	require.NoError(t, err)
	fsmSnap := snap.(*fsmSnapshot)

	require.Equalf(t, uint64(len(events)), fsmSnap.BalloonVersion, "The balloon version should match")
	// INFO!! the seqnum gets increased in the number of column families affected by a transaction
	require.Equalf(t, uint64(4), fsmSnap.LastSeqNum, "The lastSeqNum should be only 1")
	require.Equalf(t, node.ClusterInfo(), fsmSnap.Info, "The cluster info should match")
}

func TestRestore(t *testing.T) {

	log.SetLogger("TestSnapshot", log.SILENT)

	// start only one seed
	node, _, err := newSeed(1)
	require.NoError(t, err)

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node.IsLeader), "a single node is not leader!")

	// add some events
	events := append([][]byte{},
		[]byte("The year’s at the spring,"),
		[]byte("And day's at the morn;"),
		[]byte("Morning's at seven;"),
		[]byte("The hill-side’s dew-pearled;"),
	)

	// apply to WAL
	_, err = node.AddBulk(events)
	require.NoError(t, err)

	// force a raft snapshot
	f := node.raft.Snapshot()
	require.NoError(t, f.Error())

	// stop the node
	require.NoError(t, node.Close(true))

	// restart the node
	node, clean, err := newSeed(1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node.Close(true))
		clean()
	}()

	// take a snapshot to inspect its values
	snap, err := node.Snapshot()
	require.NoError(t, err)
	fsmSnap := snap.(*fsmSnapshot)

	require.Equalf(t, uint64(len(events)), fsmSnap.BalloonVersion, "The balloon version should match")
	require.Equalf(t, uint64(16), fsmSnap.LastSeqNum, "The lastSeqNum should be only 1")
	require.Equalf(t, node.ClusterInfo(), fsmSnap.Info, "The cluster info should match")
}
