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

package raftwal

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/bbva/qed/protocol"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage/badger"
	utilrand "github.com/bbva/qed/testutils/rand"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
)

func init() {
	log.SetLogger("testRaft", log.DEBUG)
}

func raftAddr(id int) string {
	return fmt.Sprintf(":830%d", id)
}
func joinAddr(id int) string {
	return fmt.Sprintf(":840%d", id)
}

func newNode(t *testing.T, id int) (*RaftBalloon, func()) {
	badgerPath := fmt.Sprintf("/var/tmp/raft-test/node%d/badger", id)

	os.MkdirAll(badgerPath, os.FileMode(0755))
	badger, err := badger.NewBadgerStore(badgerPath)
	require.NoError(t, err)

	raftPath := fmt.Sprintf("/var/tmp/raft-test/node%d/raft", id)
	os.MkdirAll(raftPath, os.FileMode(0755))
	r, err := NewRaftBalloon(raftPath, raftAddr(id), fmt.Sprintf("%d", id), badger, make(chan *protocol.Snapshot, 100))
	require.NoError(t, err)

	return r, func() {
		fmt.Println("Removing node folder")
		os.RemoveAll(fmt.Sprintf("/var/tmp/raft-test/node%d", id))
	}
}

func Test_Raft_IsLeader(t *testing.T) {

	log.SetLogger("Test_Raft_IsLeader", log.SILENT)

	r, clean := newNode(t, 1)
	defer clean()

	err := r.Open(true)
	require.NoError(t, err)

	defer func() {
		err = r.Close(true)
		require.NoError(t, err)
	}()

	_, err = r.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	require.True(t, r.IsLeader(), "single node is not leader!")

}

func Test_Raft_OpenStoreCloseSingleNode(t *testing.T) {

	r, clean := newNode(t, 2)
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

	log.SetLogger("Test_Raft_MultiNodeJoin", log.SILENT)

	r0, clean0 := newNode(t, 3)
	defer func() {
		err := r0.Close(true)
		require.NoError(t, err)
		clean0()
	}()

	err := r0.Open(true)
	require.NoError(t, err)

	_, err = r0.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	r1, clean1 := newNode(t, 4)
	defer func() {
		err := r1.Close(true)
		require.NoError(t, err)
		clean1()
	}()

	err = r1.Open(false)
	require.NoError(t, err)

	err = r0.Join("1", string(r1.raft.transport.LocalAddr()))
	require.NoError(t, err)

}

func Test_Raft_MultiNodeJoinRemove(t *testing.T) {

	r0, clean0 := newNode(t, 5)
	defer func() {
		err := r0.Close(true)
		require.NoError(t, err)
		clean0()
	}()

	err := r0.Open(true)
	require.NoError(t, err)

	_, err = r0.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	r1, clean1 := newNode(t, 6)
	defer func() {
		err := r1.Close(true)
		require.NoError(t, err)
		clean1()
	}()

	err = r1.Open(false)
	require.NoError(t, err)

	err = r0.Join("6", string(r1.raft.transport.LocalAddr()))
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

func Test_Raft_SingleNodeSnapshotOnDisk(t *testing.T) {
	r0, clean0 := newNode(t, 7)

	err := r0.Open(true)
	require.NoError(t, err)

	_, err = r0.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	// Add event
	rand.Seed(42)
	expectedBalloonVersion := uint64(rand.Intn(50))
	for i := uint64(0); i < expectedBalloonVersion; i++ {
		_, err = r0.Add([]byte(fmt.Sprintf("Test Event %d", i)))
		require.NoError(t, err)
	}
	// force snapshot
	// Snap the node and write to disk.
	snapshot, err := r0.fsm.Snapshot()
	require.NoError(t, err)

	snapDir := mustTempDir()
	defer os.RemoveAll(snapDir)
	snapFile, err := os.Create(filepath.Join(snapDir, "snapshot"))
	require.NoError(t, err)

	sink := &mockSnapshotSink{snapFile}
	err = snapshot.Persist(sink)
	require.NoError(t, err)

	// Check restoration.
	snapFile, err = os.Open(filepath.Join(snapDir, "snapshot"))
	require.NoError(t, err)

	err = r0.Close(true)
	require.NoError(t, err)
	clean0()

	r8, clean8 := newNode(t, 8)
	defer func() {
		err = r8.Close(true)
		require.NoError(t, err)
		clean8()
	}()

	err = r8.Open(true)
	require.NoError(t, err)

	_, err = r8.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	err = r8.fsm.Restore(snapFile)
	require.NoError(t, err)

	require.Equal(t, expectedBalloonVersion, r8.fsm.balloon.Version(), "Error in state recovery from snapshot")

}

func Test_Raft_SingleNodeSnapshotConsistency(t *testing.T) {
	r0, clean0 := newNode(t, 8)

	err := r0.Open(true)
	require.NoError(t, err)

	_, err = r0.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	// Add event
	rand.Seed(42)
	expectedBalloonVersion := uint64(9000)

	var wg, wgSnap sync.WaitGroup
	var snapshot raft.FSMSnapshot

	wg.Add(1)
	wgSnap.Add(1)
	go func() {
		defer wg.Done()
		for i := uint64(0); i < 20000; i++ {
			if i == expectedBalloonVersion {
				// force snapshot
				// Snap the node and write to disk.
				snapshot, err = r0.fsm.Snapshot()
				require.NoError(t, err)
				wgSnap.Done()
			}
			_, err = r0.Add([]byte(fmt.Sprintf("Test Event %d", i)))
			require.NoError(t, err)
		}
	}()

	wgSnap.Wait()

	snapDir := mustTempDir()
	// defer os.RemoveAll(snapDir)
	snapFile, err := os.Create(filepath.Join(snapDir, "snapshot"))
	require.NoError(t, err)

	sink := &mockSnapshotSink{snapFile}
	err = snapshot.Persist(sink)
	require.NoError(t, err)

	// Check restoration.
	snapFile, err = os.Open(filepath.Join(snapDir, "snapshot"))
	require.NoError(t, err)

	wg.Wait()
	err = r0.Close(true)
	require.NoError(t, err)
	clean0()

	r9, clean9 := newNode(t, 9)
	defer func() {
		err = r9.Close(true)
		require.NoError(t, err)
		clean9()
	}()

	err = r9.Open(true)
	require.NoError(t, err)

	_, err = r9.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	err = r9.fsm.Restore(snapFile)
	require.NoError(t, err)

	require.Equal(t, expectedBalloonVersion, r9.fsm.balloon.Version(), "Error in state recovery from snapshot")

}

type mockSnapshotSink struct {
	*os.File
}

func (m *mockSnapshotSink) ID() string {
	return "1"
}

func (m *mockSnapshotSink) Cancel() error {
	return nil
}

func mustTempDir() string {
	var err error
	path, err := ioutil.TempDir("", "raft-test-")
	if err != nil {
		panic("failed to create temp dir")
	}
	return path
}

func newNodeBench(b *testing.B, id int) (*RaftBalloon, func()) {
	badgerPath := fmt.Sprintf("/var/tmp/raft-test/node%d/badger", id)

	os.MkdirAll(badgerPath, os.FileMode(0755))
	badger, err := badger.NewBadgerStore(badgerPath)
	require.NoError(b, err)

	raftPath := fmt.Sprintf("/var/tmp/raft-test/node%d/raft", id)
	os.MkdirAll(raftPath, os.FileMode(0755))
	r, err := NewRaftBalloon(raftPath, raftAddr(id), fmt.Sprintf("%d", id), badger, make(chan *protocol.Snapshot, 100))
	require.NoError(b, err)

	return r, func() {
		fmt.Println("Removing node folder")
		os.RemoveAll(fmt.Sprintf("/var/tmp/raft-test/node%d", id))
	}

}

func BenchmarkRaftAdd(b *testing.B) {

	log.SetLogger("BenchmarkRaftAdd", log.SILENT)

	r, clean := newNodeBench(b, 1)
	defer clean()

	err := r.Open(true)
	require.NoError(b, err)

	b.ResetTimer()
	// b.N shoul be eq or greater than 500k to avoid benchmark framework spreding more than one goroutine.
	b.N = 500000
	nilCount := 0
	notNilCount := 0
	for i := 0; i < b.N; i++ {
		event := utilrand.Bytes(128)
		comm, _ := r.Add(event)
		if comm == nil {
			nilCount++
		} else {
			notNilCount++
		}
	}
	fmt.Printf("Nil: %d, Not Nil: %d\n", nilCount, notNilCount)
}
