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
	"fmt"
	"testing"
	"time"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/storage/rocks"
	utilrand "github.com/bbva/qed/testutils/rand"
	"github.com/stretchr/testify/require"
)

func TestSnapshot(t *testing.T) {

	// start only one seed
	node, clean, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node.Close(true))
		clean(true)
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
}

func TestRestoreSingleNode(t *testing.T) {
	// t.Skip()

	// start only one seed
	node, clean0, err := newSeed(t.Name(), 1)
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
	clean0(false)

	// restart the node
	node, clean, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node.Close(true))
		clean(true)
	}()

	// take a snapshot to inspect its values
	snap, err := node.Snapshot()
	require.NoError(t, err)
	fsmSnap := snap.(*fsmSnapshot)

	require.Equalf(t, uint64(len(events)), fsmSnap.BalloonVersion, "The balloon version should match")
	require.Equalf(t, uint64(16), fsmSnap.LastSeqNum, "The lastSeqNum should be only 1")

}

func TestRestoreFromLeaderWAL(t *testing.T) {

	// start only one seed
	node1, clean1, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node1.Close(true))
		clean1(true)
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node1.IsLeader), "a single node is not leader!")

	//start two nodes and join the cluster
	node2, clean2, err := newFollower(t.Name(), 2, node1.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node2.Close(true))
		clean2(true)
	}()

	node3, clean3, err := newFollower(t.Name(), 3, node1.info.RaftAddr)
	require.NoError(t, err)

	// check number of nodes in the cluster
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	// stop the follower
	require.NoError(t, node3.Close(true))
	clean3(false)

	// write some events
	events := append([][]byte{},
		[]byte("The year’s at the spring,"),
		[]byte("And day's at the morn;"),
		[]byte("Morning's at seven;"),
		[]byte("The hill-side’s dew-pearled;"),
	)
	_, err = node1.AddBulk(events)
	require.NoError(t, err)

	// restart follower
	node3, clean3, err = newFollower(t.Name(), 3)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node3.Close(true))
		clean3(true)
	}()

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	//wait for WAL replication
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return node1.state.Index == node3.state.Index
		}), "WAL not in sync")

	// take a snapshot to inspect its values
	snap1, err := node1.Snapshot()
	require.NoError(t, err)
	snap3, err := node3.Snapshot()
	require.NoError(t, err)
	fsmSnap1 := snap1.(*fsmSnapshot)
	fsmSnap3 := snap3.(*fsmSnapshot)

	require.Equalf(t, fsmSnap1.BalloonVersion, fsmSnap3.BalloonVersion, "The balloon version should match")
	require.Equalf(t, fsmSnap1.LastSeqNum, fsmSnap3.LastSeqNum, "The lastSeqNum should be only 1")
}

func TestRestoreFromLeaderSnapshot(t *testing.T) {
	// t.Skip()

	// start only one seed
	node1, clean1, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node1.Close(true))
		clean1(true)
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node1.IsLeader), "a single node is not leader!")

	//start two nodes and join the cluster
	node2, clean2, err := newFollower(t.Name(), 2, node1.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node2.Close(true))
		clean2(true)
	}()
	node3, clean3, err := newFollower(t.Name(), 3, node1.info.RaftAddr)
	require.NoError(t, err)

	// check number of nodes in the cluster
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	// stop the follower
	require.NoError(t, node3.Close(true))
	clean3(false)

	// write some events
	events := append([][]byte{},
		[]byte("The year’s at the spring,"),
		[]byte("And day's at the morn;"),
		[]byte("Morning's at seven;"),
		[]byte("The hill-side’s dew-pearled;"),
	)
	_, err = node1.AddBulk(events)
	require.NoError(t, err)

	// force a raft snapshot in both nodes in order to truncate the WAL
	require.NoError(t, node1.raft.Snapshot().Error())
	require.NoError(t, node2.raft.Snapshot().Error())

	// restart follower
	node3, clean3, err = newFollower(t.Name(), 3)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node3.Close(true))
		clean3(true)
	}()

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	//wait for WAL replication
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return node1.state.Index == node3.state.Index
		}), "WAL not in sync")

	// take a snapshot to inspect its values
	snap1, err := node1.Snapshot()
	require.NoError(t, err)
	snap3, err := node3.Snapshot()
	require.NoError(t, err)
	fsmSnap1 := snap1.(*fsmSnapshot)
	fsmSnap3 := snap3.(*fsmSnapshot)

	require.Equalf(t, fsmSnap1.BalloonVersion, fsmSnap3.BalloonVersion, "The balloon version should match")
	require.Equalf(t, fsmSnap1.LastSeqNum, fsmSnap3.LastSeqNum, "The lastSeqNum should be only 1")

}

func TestRestoreNewNodeFromLeader(t *testing.T) {

	// start only one seed
	node1, clean1, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node1.Close(true))
		clean1(true)
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node1.IsLeader), "a single node is not leader!")

	// start another node and join the cluster
	node2, clean2, err := newFollower(t.Name(), 2, node1.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node2.Close(true))
		clean2(true)
	}()

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 2
		}), "The number of nodes does not match")

	// write some events
	events := append([][]byte{},
		[]byte("The year’s at the spring,"),
		[]byte("And day's at the morn;"),
		[]byte("Morning's at seven;"),
		[]byte("The hill-side’s dew-pearled;"),
	)
	_, err = node1.AddBulk(events)
	require.NoError(t, err)

	// force a raft snapshot in both nodes in order to truncate the WAL
	require.NoError(t, node1.raft.Snapshot().Error())
	require.NoError(t, node2.raft.Snapshot().Error())

	// start a third node
	node3, clean3, err := newFollower(t.Name(), 3, node1.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node3.Close(true))
		clean3(true)
	}()

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	// wait for WAL replication
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return node1.state.Index == node3.state.Index
		}), "WAL not in sync")

	// take a snapshot to inspect its values
	snap1, err := node1.Snapshot()
	require.NoError(t, err)
	snap3, err := node3.Snapshot()
	require.NoError(t, err)
	fsmSnap1 := snap1.(*fsmSnapshot)
	fsmSnap3 := snap3.(*fsmSnapshot)

	require.Equalf(t, fsmSnap1.BalloonVersion, fsmSnap3.BalloonVersion, "The balloon version should match")
	require.Equalf(t, fsmSnap1.LastSeqNum, fsmSnap3.LastSeqNum, "The lastSeqNum should be only 1")

}

func TestRestoreFailureDueToGap(t *testing.T) {

	t.Skip()

	// start only one seed
	opts := DefaultClusteringOptions()
	opts.NodeID = fmt.Sprintf("%s_%d", t.Name(), 1)
	opts.Addr = raftAddr(1)
	opts.MgmtAddr = mgmtAddr(1)
	opts.HttpAddr = httpAddr(1)
	opts.Bootstrap = true
	opts.SnapshotThreshold = 0
	opts.TrailingLogs = 0
	rocksOpts := rocks.DefaultOptions()
	rocksOpts.MaxTotalWalSize = 1 * 1024 * 1024 // set to only 1mb to aggressively flush write buffers
	rocksOpts.WALSizeLimitMB = 0                // force to truncate the rocksdb WAL
	rocksOpts.WALTtlSeconds = 1
	node1, clean1, err := newNode(opts, rocksOpts)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node1.Close(true))
		clean1(true)
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node1.IsLeader), "a single node is not leader!")

	// write some events
	for i := uint64(0); i < 30000; i++ {
		event := utilrand.Bytes(128)
		_, err = node1.Add(event)
		require.NoError(t, err)
	}

	// force a raft snapshot in both nodes in order to truncate the raft WAL
	require.NoError(t, node1.raft.Snapshot().Error())

	// start another node and join the cluster
	_, _, err = newFollower(t.Name(), 2, node1.info.RaftAddr)
	require.Error(t, err)

}

func TestRestoreNewNodeFromChangedLeader(t *testing.T) {

	// start only one seed
	node1, clean1, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node1.Close(true))
		clean1(true)
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node1.IsLeader), "a single node is not leader!")

	// start another node and join the cluster
	node2, clean2, err := newFollower(t.Name(), 2, node1.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node2.Close(true))
		clean2(true)
	}()

	// write some events
	for i := uint64(0); i < 10000; i++ {
		event := utilrand.Bytes(128)
		_, err = node1.Add(event)
		require.NoError(t, err)
	}

	// force a raft snapshot in both nodes in order to truncate the WAL
	require.NoError(t, node1.raft.Snapshot().Error())
	require.NoError(t, node2.raft.Snapshot().Error())

	// start a third node
	node3, clean3, err := newFollower(t.Name(), 3, node1.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node3.Close(true))
		clean3(true)
	}()

	// change leader
	node1.leaveLeadership()

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	// wait for WAL replication
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return node1.state.Index == node3.state.Index
		}), "WAL not in sync")

	// take a snapshot to inspect its values
	snap1, err := node1.Snapshot()
	require.NoError(t, err)
	snap3, err := node3.Snapshot()
	require.NoError(t, err)
	fsmSnap1 := snap1.(*fsmSnapshot)
	fsmSnap3 := snap3.(*fsmSnapshot)

	require.Equalf(t, fsmSnap1.BalloonVersion, fsmSnap3.BalloonVersion, "The balloon version should match")
	require.Equalf(t, fsmSnap1.LastSeqNum, fsmSnap3.LastSeqNum, "The lastSeqNum should be only 1")

}

func TestRestoreOldNodeFromChangedLeader(t *testing.T) {

	// start only one seed
	node1, clean1, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node1.Close(true))
		clean1(true)
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node1.IsLeader), "a single node is not leader!")

	// start another two nodes and join the cluster
	node2, clean2, err := newFollower(t.Name(), 2, node1.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node2.Close(true))
		clean2(true)
	}()
	node3, clean3, err := newFollower(t.Name(), 3, node1.info.RaftAddr)
	require.NoError(t, err)

	// check number of nodes in the cluster
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	// wait for WAL replication
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return node1.state.Index == node3.state.Index
		}), "WAL not in sync")

	// stop the follower
	require.NoError(t, node3.Close(true))
	clean3(false)

	// write some events
	for i := uint64(0); i < 10000; i++ {
		event := utilrand.Bytes(128)
		_, err = node1.Add(event)
		require.NoError(t, err)
	}

	// force a raft snapshot in both nodes in order to truncate the WAL
	require.NoError(t, node1.raft.Snapshot().Error())
	require.NoError(t, node2.raft.Snapshot().Error())

	// change leader
	node1.leaveLeadership()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, node2.IsLeader), "the node is not leader!")

	// restart a third node
	node3, clean3, err = newFollower(t.Name(), 3)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node3.Close(true))
		clean3(true)
	}()

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return len(node1.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	// wait for WAL replication
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return node1.state.Index == node3.state.Index
		}), "WAL not in sync")

	// take a snapshot to inspect its values
	snap1, err := node1.Snapshot()
	require.NoError(t, err)
	snap3, err := node3.Snapshot()
	require.NoError(t, err)
	fsmSnap1 := snap1.(*fsmSnapshot)
	fsmSnap3 := snap3.(*fsmSnapshot)

	require.Equalf(t, fsmSnap1.BalloonVersion, fsmSnap3.BalloonVersion, "The balloon version should match")
	require.Equalf(t, fsmSnap1.LastSeqNum, fsmSnap3.LastSeqNum, "The lastSeqNum should be only 1")

}
