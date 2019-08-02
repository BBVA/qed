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

	raftbench "github.com/hashicorp/raft/bench"
)

func BenchmarkRaftLogFirstIndex(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.FirstIndex(b, store)
}

func BenchmarkRaftLogLastIndex(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.LastIndex(b, store)
}

func BenchmarkRaftLogGetLog(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.GetLog(b, store)
}

func BenchmarkRaftLogStoreLog(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.StoreLog(b, store)
}

func BenchmarkRaftLogStoreLogs(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.StoreLogs(b, store)
}

func BenchmarkRaftLogDeleteRange(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.DeleteRange(b, store)
}

func BenchmarkRaftLogSet(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.Set(b, store)
}

func BenchmarkRaftLogGet(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.Get(b, store)
}

func BenchmarkRaftLogSetUint64(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.SetUint64(b, store)
}

func BenchmarkRaftLogGetUint64(b *testing.B) {
	store, _, closeF := openRaftLog(b)
	defer closeF()

	raftbench.GetUint64(b, store)
}
