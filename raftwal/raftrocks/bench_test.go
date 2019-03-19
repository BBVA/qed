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
	"os"
	"testing"

	"github.com/hashicorp/raft/bench"
)

func BenchmarkRocksDBStore_FirstIndex(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.FirstIndex(b, store)
}

func BenchmarkRocksDBStore_LastIndex(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.LastIndex(b, store)
}

func BenchmarkRocksDBStore_GetLog(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.GetLog(b, store)
}

func BenchmarkRocksDBStore_StoreLog(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.StoreLog(b, store)
}

func BenchmarkRocksDBStore_StoreLogs(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.StoreLogs(b, store)
}

func BenchmarkRocksDBStore_DeleteRange(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.DeleteRange(b, store)
}

func BenchmarkRocksDBStore_Set(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.Set(b, store)
}

func BenchmarkRocksDBStore_Get(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.Get(b, store)
}

func BenchmarkRocksDBStore_SetUint64(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.SetUint64(b, store)
}

func BenchmarkRocksDBStore_GetUint64(b *testing.B) {
	store, path := testRocksDBStore(b)
	defer store.Close()
	defer os.Remove(path)

	raftbench.GetUint64(b, store)
}
