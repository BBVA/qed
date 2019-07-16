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
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIterator(t *testing.T) {

	db, path := newTestDB(t, "TestIterator", nil)
	defer func() {
		db.Close()
		os.RemoveAll(path)
	}()

	// insert keys
	givenKeys := [][]byte{[]byte("key1"), []byte("key2"), []byte("key3")}
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		require.NoError(t, db.Put(wo, k, []byte("val")))
	}

	ro := NewDefaultReadOptions()
	iter := db.NewIterator(ro)
	defer iter.Close()

	var actualKeys [][]byte
	for iter.SeekToFirst(); iter.Valid(); iter.Next() {
		key := make([]byte, 4)
		copy(key, iter.Key().Data())
		actualKeys = append(actualKeys, key)
	}
	require.NoError(t, iter.Err())
	require.Equal(t, actualKeys, givenKeys)
}
