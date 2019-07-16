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

	log.SetLogger("TestOpenAndCloseRaftNode", log.SILENT)

	// start only one seed
	r, clean, err := newSeed(t, 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close(true))
		clean()
	}()

}

func TestRaftNodeIsLeader(t *testing.T) {

	log.SetLogger("TestRaftNodeIsLeader", log.SILENT)

	// start only one seed
	r, clean, err := newSeed(t, 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close(true))
		clean()
	}()

	// check the leader of the cluster
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, r.IsLeader), "a single node is not leader!")

}

func TestRaftNodeNotLeader(t *testing.T) {

	log.SetLogger("TestRaftNodeNotLeader", log.SILENT)

	// start only one follower
	_, clean, err := newFollower(t, 1)
	require.Error(t, err)
	defer func() {
		clean()
	}()

}

func TestRaftNodeClusterInfo(t *testing.T) {

	log.SetLogger("TestRaftNodeClusterInfo", log.SILENT)

	// start only one seed
	r, clean, err := newSeed(t, 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r.Close(true))
		clean()
	}()

	// check cluster info
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, func() bool {
			return len(r.ClusterInfo().Nodes) == 1
		}), "The number of nodes does not match")
	require.Truef(t,
		retryTrue(50, 200*time.Millisecond, func() bool {
			return r.Info().NodeId == r.ClusterInfo().LeaderId
		}), "The leaderId in cluster info is correct")

}

func TestMultiRaftNodeJoin(t *testing.T) {

	log.SetLogger("TestMultiRaftNodeJoin", log.SILENT)

	// start one seed
	r0, clean0, err := newSeed(t, 0)
	require.NoError(t, err)

	require.Truef(t, retryTrue(50, 200*time.Millisecond, r0.IsLeader), "a single node is not leader!")

	// start one follower and join the cluster
	r1, clean1, err := newFollower(t, 1, r0.info.ClusterMgmtAddr)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, r0.Close(true))
		require.NoError(t, r1.Close(true))
		clean0()
		clean1()
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

	log.SetLogger("TestMultiRaftNodeJoin", log.SILENT)

	// start one seed
	r0, clean0, err := newSeed(t, 0)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r0.Close(true))
		clean0()
	}()

	// check leader
	require.Truef(t, retryTrue(50, 200*time.Millisecond, r0.IsLeader), "r0 is not leader!")

	// star one follower and join the cluster
	r1, clean1, err := newFollower(t, 1, r0.info.ClusterMgmtAddr)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, r1.Close(true))
		clean1()
	}()

	// check number of nodes in the cluster
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(r0.ClusterInfo().Nodes) == 2
		}), "The number of nodes does not match")

	// start another follower but try to join to a non-leader node
	_, clean2, err := newFollower(t, 2, r1.info.ClusterMgmtAddr) // wrong address
	defer clean2()
	require.Error(t, err)

	// check the number of nodes in the cluster
	require.Truef(t,
		retryTrue(10, 200*time.Millisecond, func() bool {
			return len(r0.ClusterInfo().Nodes) == 2
		}), "The number of nodes does not match")

}

type closeF func()

func raftAddr(id int) string {
	return fmt.Sprintf(":1830%d", id)
}

func clusterMgmtAddr(id int) string {
	return fmt.Sprintf(":1930%d", id)
}

func newSeed(t *testing.T, id int) (*RaftNode, closeF, error) {
	return newNode(t, id, true)
}

func newFollower(t *testing.T, id int, seeds ...string) (*RaftNode, closeF, error) {
	return newNode(t, id, false, seeds...)
}

func newNode(t *testing.T, id int, bootstrap bool, seeds ...string) (*RaftNode, closeF, error) {

	dbPath := fmt.Sprintf("/var/tmp/cluster-test/node%d/db", id)
	err := os.MkdirAll(dbPath, os.FileMode(0755))
	require.NoError(t, err)
	db, err := rocks.NewRocksDBStore(dbPath)
	require.NoError(t, err)

	raftPath := fmt.Sprintf("/var/tmp/cluster-test/node%d/raft", id)
	err = os.MkdirAll(raftPath, os.FileMode(0755))
	require.NoError(t, err)

	opts := DefaultClusteringOptions()
	opts.NodeID = fmt.Sprintf("%d", id)
	opts.Addr = raftAddr(id)
	opts.ClusterMgmtAddr = clusterMgmtAddr(id)
	opts.RaftLogPath = raftPath
	opts.Bootstrap = bootstrap
	opts.Seeds = seeds
	r, err := NewRaftNode(opts, db, make(chan *protocol.Snapshot, 25000))

	return r, func() {
		os.RemoveAll(fmt.Sprintf("/var/tmp/cluster-test/node%d", id))
	}, err

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
