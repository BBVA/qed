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
	"os"
	"testing"
	"time"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/storage/rocks"
	"github.com/stretchr/testify/require"
)

func TestOpenAndCloseRaftNode(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start only one seed
	r, clean, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close(true))
		clean(true)
	}()

}

func TestRaftNodeIsLeader(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start only one seed
	r, clean, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close(true))
		clean(true)
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, r.IsLeader), "a single node is not leader!")

}

func TestRaftNodeNotLeader(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start only one follower
	_, clean, err := newFollower(t.Name(), 1)
	require.Error(t, err)
	defer func() {
		clean(true)
	}()

}

func TestRaftNodeClusterInfo(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start only one seed
	r, clean, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close(true))
		clean(true)
	}()

	// check cluster info
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, func() bool {
			return len(r.ClusterInfo().Nodes) == 1
		}), "The number of nodes does not match")

	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, func() bool {
			ci := r.ClusterInfo()
			ni := r.Info()
			return ci.LeaderId == ni.NodeId
		}), "The leaderId in cluster info is correct")

}

func TestMultiRaftNodeJoin(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start one seed
	r0, clean0, err := newSeed(t.Name(), 0)
	require.NoError(t, err)

	require.Truef(t, retryTrue(50, 200*time.Millisecond, r0.IsLeader), "a single node is not leader!")

	// start one follower and join the cluster
	r1, clean1, err := newFollower(t.Name(), 1, r0.info.RaftAddr)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, r0.Close(true))
		require.NoError(t, r1.Close(true))
		clean0(true)
		clean1(true)
	}()

	// check cluster info
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(r0.ClusterInfo().Nodes) == 2
		}), "The number of nodes does not match")
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(r1.ClusterInfo().Nodes) == 2
		}), "The number of nodes does not match")

}

func TestMultiRaftNodesJoinNotLeader(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start one seed
	r0, clean0, err := newSeed(t.Name(), 0)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r0.Close(true))
		clean0(true)
	}()

	// check leader
	require.Truef(t, retryTrue(50, 200*time.Millisecond, r0.IsLeader), "r0 is not leader!")

	// star one follower and join the cluster
	r1, clean1, err := newFollower(t.Name(), 1, r0.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r1.Close(true))
		clean1(true)
	}()

	// check number of nodes in the cluster
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(r0.ClusterInfo().Nodes) == 2
		}), "The number of nodes does not match")

	// start another follower but try to join to a non-leader node
	_, clean2, err := newFollower(t.Name(), 2, r1.info.RaftAddr) // wrong address
	defer clean2(true)
	require.Error(t, err)

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(r0.ClusterInfo().Nodes) == 2
		}), "The number of nodes does not match")

}

func TestMultRaftNodesReJoin(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start one seed
	r0, clean0, err := newSeed(t.Name(), 0)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r0.Close(true))
		clean0(true)
	}()

	// check leader
	require.Truef(t, retryTrue(50, 200*time.Millisecond, r0.IsLeader), "r0 is not leader!")

	// start two replicas
	r1, clean1, err := newFollower(t.Name(), 1, r0.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r1.Close(true))
		clean1(true)
	}()
	r2, clean2, err := newFollower(t.Name(), 2, r0.info.RaftAddr)
	require.NoError(t, err)

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, func() bool {
			return len(r2.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	time.Sleep(1 * time.Second)

	// stop one node
	r2.Close(true)
	clean2(false)

	// restart the stopped node
	r2, clean3, err := newFollower(t.Name(), 2, r0.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r2.Close(true))
		clean3(true)
	}()

	time.Sleep(1 * time.Second)

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(20, 200*time.Millisecond, func() bool {
			return len(r0.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

}

func TestMultiRaftLeaveLeadership(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start one seed
	r0, clean0, err := newSeed(t.Name(), 0)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r0.Close(true))
		clean0(true)
	}()

	// check leader
	require.Truef(t, retryTrue(50, 200*time.Millisecond, r0.IsLeader), "r0 is not leader!")

	// start two replicas
	r1, clean1, err := newFollower(t.Name(), 1, r0.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r1.Close(true))
		clean1(true)
	}()
	r2, clean2, err := newFollower(t.Name(), 2, r0.info.RaftAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r2.Close(true))
		clean2(true)
	}()

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, func() bool {
			return len(r0.ClusterInfo().Nodes) == 3
		}), "The number of nodes does not match")

	// leave leadership
	require.NoError(t, r0.leaveLeadership())

	time.Sleep(3 * time.Second)

	// check
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, func() bool {
			return r0.ClusterInfo().LeaderId != r0.Info().NodeId && r0.ClusterInfo().LeaderId != ""
		}), "The leader has to have changed")

}

type closeF func(dir bool)

func raftAddr(id int) string {
	return fmt.Sprintf("127.0.0.1:1830%d", id)
}

func mgmtAddr(id int) string {
	return fmt.Sprintf("127.0.0.1:1930%d", id)
}

func httpAddr(id int) string {
	return fmt.Sprintf("127.0.0.1:1730%d", id)
}

func newSeed(name string, id int) (*RaftNode, closeF, error) {
	opts := DefaultClusteringOptions()
	opts.NodeID = fmt.Sprintf("%s_%d", name, id)
	opts.Addr = raftAddr(id)
	opts.MgmtAddr = mgmtAddr(id)
	opts.HttpAddr = httpAddr(id)
	opts.Bootstrap = true
	opts.SnapshotThreshold = 0
	opts.TrailingLogs = 0
	rocksOpts := rocks.DefaultOptions()
	return newNode(opts, rocksOpts)
}

func newFollower(name string, id int, seeds ...string) (*RaftNode, closeF, error) {
	opts := DefaultClusteringOptions()
	opts.NodeID = fmt.Sprintf("%s_%d", name, id)
	opts.Addr = raftAddr(id)
	opts.MgmtAddr = mgmtAddr(id)
	opts.HttpAddr = httpAddr(id)
	opts.Bootstrap = false
	opts.SnapshotThreshold = 0
	opts.TrailingLogs = 0
	opts.Seeds = seeds
	rocksOpts := rocks.DefaultOptions()
	return newNode(opts, rocksOpts)
}

func newNode(opts *ClusteringOptions, rocksOpts *rocks.Options) (*RaftNode, closeF, error) {

	snapshotsCh := make(chan *protocol.Snapshot, 25000)
	snapshotsDrainer(snapshotsCh)

	// var metricsCloseF = func() {}

	cleanF := func(dir bool) {
		// metricsCloseF()
		close(snapshotsCh)
		if dir {
			os.RemoveAll(fmt.Sprintf("/var/tmp/cluster-test/node_%s", opts.NodeID))
		}
	}

	dbPath := fmt.Sprintf("/var/tmp/cluster-test/node_%s/db", opts.NodeID)
	if err := os.MkdirAll(dbPath, os.FileMode(0755)); err != nil {
		return nil, cleanF, err
	}
	rocksOpts.Path = dbPath

	db, err := rocks.NewRocksDBStoreWithOpts(rocksOpts)
	if err != nil {
		return nil, cleanF, err
	}

	raftPath := fmt.Sprintf("/var/tmp/cluster-test/node_%s/raft", opts.NodeID)
	if err := os.MkdirAll(raftPath, os.FileMode(0755)); err != nil {
		return nil, cleanF, err
	}
	opts.RaftLogPath = raftPath

	node, err := NewRaftNode(opts, db, snapshotsCh)
	if err != nil {
		return nil, cleanF, err
	}

	// metricsCloseF = metrics_utils.StartMetricsServer(node, db)

	return node, cleanF, err

}

func snapshotsDrainer(snapshotsCh chan *protocol.Snapshot) {
	go func() {
		for {
			_, ok := <-snapshotsCh
			if !ok {
				return
			}
		}
	}()
}

func retryTrue(tries int, delay time.Duration, fn func() bool) bool {
	var i int
	var exit bool
	for i = 0; i < tries; i++ {
		exit = fn()
		if exit {
			break
		}
		time.Sleep(delay)
	}
	return exit
}

func retryErr(tries int, delay time.Duration, fn func() error) error {
	var i int
	var err error
	for i = 0; i < tries; i++ {
		err = fn()
		if err == nil {
			break
		}
		time.Sleep(delay)
	}
	return err
}
