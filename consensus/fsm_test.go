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
	"github.com/bbva/qed/testutils/rand"
	utilrand "github.com/bbva/qed/testutils/rand"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
)

func TestApplyAdd(t *testing.T) {

	log.SetLogger(t.Name(), log.SILENT)

	// start only one seed
	node, clean, err := newSeed(t.Name(), 1)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, node.Close(true))
		clean(true)
	}()

	require.Truef(t, retryTrue(50, 200*time.Millisecond, node.IsLeader), "a single node is not leader!")

	h := hashing.NewSha256Hasher()
	event := append([]hashing.Digest{},
		h.Do([]byte("The year’s at the spring,")),
		h.Do([]byte("And day's at the morn;")),
		h.Do([]byte("Morning's at seven;")),
		h.Do([]byte("The hill-side’s dew-pearled;")),
		h.Do([]byte("The lark's on the wing;")),
		h.Do([]byte("The snail's on the thorn;")),
		h.Do([]byte("God's in his heaven—")),
		h.Do([]byte("All's right with the world!")),
	)
	cmd := newCommand(addEventCommandType)
	cmd.encode(event)

	tests := []struct {
		log           *raft.Log
		expectedError bool
	}{
		{newLog(1, 1, cmd.data), false}, // happy path
		{newLog(1, 1, cmd.data), true},  // Error: Command already applied
		{newLog(2, 1, cmd.data), false}, // happy path
		{newLog(1, 1, cmd.data), true},  // Error: Command out of order
	}

	for i, test := range tests {
		r := node.Apply(test.log).(*fsmResponse)
		require.Equalf(t, test.expectedError, r.err != nil, "failed in test case %d", i)
	}
}

func BenchmarkApplyAdd(b *testing.B) {

	log.SetLogger(b.Name(), log.SILENT)

	// start only one seed
	node, clean, err := newSeed(b.Name(), 1)
	require.NoError(b, err)
	defer func() {
		require.NoError(b, node.Close(true))
		clean(true)
	}()

	require.Truef(b, retryTrue(50, 200*time.Millisecond, node.IsLeader), "a single node is not leader!")

	hasher := hashing.NewSha256Hasher()
	b.ResetTimer()
	b.N = 2000000
	for i := 0; i < b.N; i++ {
		cmd := newCommand(addEventCommandType)
		cmd.encode([]hashing.Digest{hasher.Do(rand.Bytes(128))})
		log := newLog(uint64(i), uint64(1), cmd.data)
		resp := node.Apply(log)
		require.NoError(b, resp.(*fsmResponse).err)
	}

}

func BenchmarkRaftAdd(b *testing.B) {

	log.SetLogger(b.Name(), log.SILENT)

	node, clean, err := newSeed(b.Name(), 1)
	require.NoError(b, err)
	defer func() {
		require.NoError(b, node.Close(true))
		clean(true)
	}()

	require.Truef(b, retryTrue(50, 200*time.Millisecond, node.IsLeader), "a single node is not leader!")

	// b.N shoul be eq or greater than 500k to avoid benchmark framework spreading more than one goroutine.
	b.N = 2000000
	b.ResetTimer()
	b.SetParallelism(100)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			event := utilrand.Bytes(128)
			_, err := node.Add(event)
			require.NoError(b, err)
		}
	})

}

func newLog(index, term uint64, command []byte) *raft.Log {
	return &raft.Log{Index: index, Term: term, Type: raft.LogCommand, Data: command}
}
