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

package rocksdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWALIterator(t *testing.T) {

	db, _ := newTestDB(t, "TestWALIterator", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		require.NoError(t, db.Put(wo, k, []byte("val")))
	}

	// check last sequence number
	lastSeqNum := db.GetLatestSequenceNumber()
	require.Equal(t, uint64(len(givenKeys)), lastSeqNum)

	// get updates from the last sequence number
	it, err := db.GetUpdatesSince(lastSeqNum)
	defer it.Close()
	require.NoError(t, err)
	require.True(t, it.Valid())

	// check batch data
	_, seqNum := it.GetBatch()
	require.Equal(t, lastSeqNum, seqNum)

}

func TestWALIteratorFromBeginning(t *testing.T) {

	db, _ := newTestDB(t, "TestWALIteratorFromBeginning", nil)
	defer db.Close()

	// insert keys
	givenKeys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		require.NoError(t, db.Put(wo, k, []byte("val")))
	}

	// check last sequence number
	lastSeqNum := db.GetLatestSequenceNumber()
	require.Equal(t, uint64(len(givenKeys)), lastSeqNum)

	// get updates from the last sequence number
	it, err := db.GetUpdatesSince(1)
	defer it.Close()
	require.NoError(t, err)

	// check batch data
	count := uint64(1)
	for ; it.Valid(); it.Next() {
		_, seqNum := it.GetBatch()
		require.Equal(t, count, seqNum)
		count++
	}

}
