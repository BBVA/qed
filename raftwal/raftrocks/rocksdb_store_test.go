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
package raftrocks

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/bbva/qed/rocksdb"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/raft"
)

func testRocksDBStore(t testing.TB) (*RocksDBStore, string) {
	path, err := ioutil.TempDir("", "raftrocks")
	require.NoError(t, err)
	os.RemoveAll(path)

	// Successfully creates and returns a store
	store, err := NewRocksDBStore(path)
	require.NoError(t, err)

	return store, path
}

func testRaftLog(idx uint64, data string) *raft.Log {
	return &raft.Log{
		Data:  []byte(data),
		Index: idx,
	}
}

func TestRocksDBStore_Implements(t *testing.T) {
	var store interface{} = &RocksDBStore{}
	if _, ok := store.(raft.StableStore); !ok {
		t.Fatalf("RocksDBStore does not implement raft.StableStore")
	}
	if _, ok := store.(raft.LogStore); !ok {
		t.Fatalf("RocksDBStore does not implement raft.LogStore")
	}
}

func TestNewRocksDBStore(t *testing.T) {

	store, path := testRocksDBStore(t)

	// Ensure the directory was created
	require.Equal(t, path, store.path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("err: %s", err)
	}

	// Close the store so we can open again
	require.NoError(t, store.Close())

	// Ensure our files were created
	opts := rocksdb.NewDefaultOptions()
	opts.SetCreateIfMissing(false)
	_, err := rocksdb.OpenDBForReadOnly(path, opts, true)
	require.NoError(t, err)

}

func TestRocksDBStore_FirstIndex(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.RemoveAll(path)

	// Should get 0 index on empty log
	idx, err := store.FirstIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(0), idx)

	// Set a mock raft log
	logs := []*raft.Log{
		testRaftLog(1, "log1"),
		testRaftLog(2, "log2"),
		testRaftLog(3, "log3"),
	}
	require.NoError(t, store.StoreLogs(logs))

	// Fetch the first Raft index
	idx, err = store.FirstIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(1), idx)
}

func TestRocksDBStore_LastIndex(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.RemoveAll(path)

	// Should get 0 index on empty log
	idx, err := store.LastIndex()
	require.NoError(t, err)
	require.Equal(t, uint64(0), idx)

	// Set a mock raft log
	logs := []*raft.Log{
		testRaftLog(1, "log1"),
		testRaftLog(2, "log2"),
		testRaftLog(3, "log3"),
	}
	require.NoError(t, store.StoreLogs(logs))

	// Fetch the last Raft index
	idx, err = store.LastIndex()
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	require.NoError(t, err)
	require.Equal(t, uint64(3), idx)
}

func TestRocksDBStore_GetLog(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.RemoveAll(path)

	log := new(raft.Log)

	// Should return an error on non-existent log
	err := store.GetLog(1, log)
	require.Equalf(t, err, raft.ErrLogNotFound, "Expected raft log not found")

	// Set a mock raft log
	logs := []*raft.Log{
		testRaftLog(1, "log1"),
		testRaftLog(2, "log2"),
		testRaftLog(3, "log3"),
	}
	require.NoError(t, store.StoreLogs(logs))

	// Should return the proper log
	require.NoError(t, store.GetLog(2, log))
	require.Equal(t, log, logs[1])
}

func TestRocksDBStore_SetLog(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.Remove(path)

	// Create the log
	log := &raft.Log{
		Data:  []byte("log1"),
		Index: 1,
	}

	// Attempt to store the log
	require.NoError(t, store.StoreLog(log))

	// Retrieve the log again
	result := new(raft.Log)
	require.NoError(t, store.GetLog(1, result))

	// Ensure the log comes back the same
	require.Equal(t, log, result)
}

func TestRocksDBStore_SetLogs(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.Remove(path)

	// Create a set of logs
	logs := []*raft.Log{
		testRaftLog(1, "log1"),
		testRaftLog(2, "log2"),
	}

	// Attempt to store the logs
	require.NoError(t, store.StoreLogs(logs))

	// Ensure we stored them all
	result1, result2 := new(raft.Log), new(raft.Log)
	require.NoError(t, store.GetLog(1, result1))
	require.Equal(t, logs[0], result1)

	require.NoError(t, store.GetLog(2, result2))
	require.Equal(t, logs[1], result2)
}

func TestRocksDBStore_DeleteRange(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.Remove(path)

	// Create a set of logs
	log1 := testRaftLog(1, "log1")
	log2 := testRaftLog(2, "log2")
	log3 := testRaftLog(3, "log3")
	logs := []*raft.Log{log1, log2, log3}

	// Attempt to store the logs
	require.NoError(t, store.StoreLogs(logs))

	// Attempt to delete a range of logs
	require.NoError(t, store.DeleteRange(1, 2))

	// Ensure the logs were deleted
	err := store.GetLog(1, new(raft.Log))
	require.Error(t, err)
	require.Equal(t, raft.ErrLogNotFound, err)

	err = store.GetLog(2, new(raft.Log))
	require.Error(t, err)
	require.Equal(t, raft.ErrLogNotFound, err)

}

func TestRocksDBStore_Set_Get(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.Remove(path)

	// Returns error on non-existent key
	_, err := store.Get([]byte("bad"))
	require.Error(t, err)
	require.Equal(t, err, ErrKeyNotFound)

	k, v := []byte("hello"), []byte("world")

	// Try to set a k/v pair
	require.NoError(t, store.Set(k, v))

	// Try to read it back
	val, err := store.Get(k)
	require.NoError(t, err)
	require.Equal(t, v, val)
}

func TestRocksDBStore_SetUint64_GetUint64(t *testing.T) {
	store, path := testRocksDBStore(t)
	defer store.Close()
	defer os.Remove(path)

	// Returns error on non-existent key
	_, err := store.GetUint64([]byte("bad"))
	require.Error(t, err)
	require.Equal(t, err, ErrKeyNotFound)

	k, v := []byte("abc"), uint64(123)

	// Attempt to set the k/v pair
	require.NoError(t, store.SetUint64(k, v))

	// Read back the value
	val, err := store.GetUint64(k)
	require.NoError(t, err)
	require.Equal(t, v, val)
}
