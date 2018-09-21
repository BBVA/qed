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

package balloon

import (
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage/badger"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLogger("testRaft", log.DEBUG)
}

func raftAddr(id int) string {
	return fmt.Sprintf("127.0.0.1:830%d", id)
}
func joinAddr(id int) string {
	return fmt.Sprintf("127.0.0.1:840%d", id)
}

func newNode(t *testing.T, id int) (*RaftBalloon, func()) {
	badgerPath := fmt.Sprintf("/var/tmp/raft-test/node%d/badger", id)

	os.MkdirAll(badgerPath, os.FileMode(0755))
	badger, err := badger.NewBadgerStore(badgerPath)
	require.NoError(t, err)

	raftPath := fmt.Sprintf("/var/tmp/raft-test/node%d/badger", id)
	os.MkdirAll(raftPath, os.FileMode(0755))
	r, err := NewRaftBalloon(raftPath, raftAddr(id), fmt.Sprintf("%d", id), badger)
	require.NoError(t, err)

	return r, func() {
		os.RemoveAll("/var/tmp/raft-test")
	}
}

func Test_Raft_IsLeader(t *testing.T) {

	r, clean := newNode(t, 0)
	defer clean()

	err := r.Open(true)
	require.NoError(t, err)

	defer r.Close(true)
	_, err = r.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	require.True(t, r.IsLeader(), "single node is not leader!")

}

func Test_Raft_OpenStoreCloseSingleNode(t *testing.T) {

	r, clean := newNode(t, 0)
	defer clean()

	err := r.Open(true)
	require.NoError(t, err)

	_, err = r.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	err = r.Close(true)
	require.NoError(t, err)

	err = r.Open(true)
	require.Equal(t, err, ErrBalloonInvalidState, err, "incorrect error returned on re-open attempt")

}

func Test_Raft_MultiNodeJoin(t *testing.T) {
	r0, clean0 := newNode(t, 0)
	defer func() {
		r0.Close(true)
		clean0()
	}()

	err := r0.Open(true)
	require.NoError(t, err)

	_, err = r0.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	r1, clean1 := newNode(t, 1)
	defer func() {
		r1.Close(true)
		clean1()
	}()

	err = r1.Open(false)
	require.NoError(t, err)

	err = r0.Join("1", string(r1.raft.transport.LocalAddr()))
	require.NoError(t, err)

}

func Test_Raft_MultiNodeJoinRemove(t *testing.T) {

	r0, clean0 := newNode(t, 0)
	defer func() {
		r0.Close(true)
		clean0()
	}()

	err := r0.Open(true)
	require.NoError(t, err)

	_, err = r0.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	r1, clean1 := newNode(t, 1)
	defer func() {
		r1.Close(true)
		clean1()
	}()

	err = r1.Open(false)
	require.NoError(t, err)

	err = r0.Join("1", string(r1.raft.transport.LocalAddr()))
	require.NoError(t, err)

	_, err = r0.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	// Check leader state on follower.
	require.Equal(t, r1.LeaderAddr(), r0.Addr(), "wrong leader address returned")

	id, err := r1.LeaderID()
	require.NoError(t, err)

	require.Equal(t, id, r0.ID(), "wrong leader ID returned")

	storeNodes := []string{r0.id, r1.id}
	sort.StringSlice(storeNodes).Sort()

	nodes, err := r0.Nodes()
	require.NoError(t, err)
	require.Equal(t, len(nodes), len(storeNodes), "size of cluster is not correct")

	if storeNodes[0] != string(nodes[0].ID) || storeNodes[1] != string(nodes[1].ID) {
		t.Fatalf("cluster does not have correct nodes")
	}

	// Remove a node.
	err = r0.Remove(r1.ID())
	require.NoError(t, err)

	nodes, err = r0.Nodes()
	require.NoError(t, err)

	require.Equal(t, len(nodes), 1, "size of cluster is not correct post remove")
	require.Equal(t, r0.ID(), string(nodes[0].ID), "cluster does not have correct nodes post remove")

}
