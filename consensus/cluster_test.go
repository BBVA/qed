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
	r, clean := newNode(t, 1, true)
	defer func() {
		require.NoError(t, r.Shutdown(true))
		clean()
	}()

}

func TestRaftNodeIsLeader(t *testing.T) {

	log.SetLogger("TestRaftNodeIsLeader", log.SILENT)

	r, clean := newNode(t, 1, true)
	defer func() {
		require.NoError(t, r.Shutdown(true))
		clean()
	}()

	require.Truef(t, retryTrue(10, 200*time.Millisecond, r.IsLeader), "a single node is not leader!")

}

func TestMultiRaftNodeJoin(t *testing.T) {

	log.SetLogger("TestMultiRaftNodeJoin", log.SILENT)

	r0, clean0 := newNode(t, 0, true)

	require.Truef(t, retryTrue(10, 200*time.Millisecond, r0.IsLeader), "a single node is not leader!")

	r1, clean1 := newNode(t, 1, false)

	defer func() {
		require.NoError(t, r0.Shutdown(true))
		require.NoError(t, r1.Shutdown(true))
		clean0()
		clean1()
	}()

	err := r1.AttemptToJoinCluster([]string{r0.Info().ClusterMgmtAddr})
	require.NoError(t, err)

}

type closeF func()

func raftAddr(id int) string {
	return fmt.Sprintf(":1830%d", id)
}

func clusterMgmtAddr(id int) string {
	return fmt.Sprintf(":1930%d", id)
}

func newNode(t *testing.T, id int, bootstrap bool) (*RaftNode, closeF) {
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
	r, err := NewRaftNode(opts, db, make(chan *protocol.Snapshot, 25000))
	require.NoError(t, err)

	return r, func() {
		os.RemoveAll(fmt.Sprintf("/var/tmp/cluster-test/node%d", id))
	}

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
