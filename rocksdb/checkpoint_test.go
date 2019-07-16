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
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckpoint(t *testing.T) {

	checkDir, err := ioutil.TempDir("", "rocksdb-checkpoint")
	require.NoError(t, err)
	err = os.RemoveAll(checkDir)
	require.NoError(t, err)

	db, path := newTestDB(t, "TestCheckpoint", nil)
	defer func() {
		db.Close()
		os.RemoveAll(path)
	}()

	// insert keys
	givenKeys := [][]byte{
		[]byte("key1"),
		[]byte("key2"),
		[]byte("key3"),
	}
	givenValue := []byte("value")
	wo := NewDefaultWriteOptions()
	for _, k := range givenKeys {
		require.NoError(t, db.Put(wo, k, givenValue))
	}

	checkpoint, err := db.NewCheckpoint()
	require.NoError(t, err)
	require.NotNil(t, checkpoint)
	defer checkpoint.Destroy()

	err = checkpoint.CreateCheckpoint(checkDir, 0)
	require.NoError(t, err)

	opts := NewDefaultOptions()
	dbCheck, err := OpenDBForReadOnly(checkDir, opts, true)
	require.NoError(t, err)
	defer dbCheck.Close()

	// test keys
	var value *Slice
	ro := NewDefaultReadOptions()
	for _, k := range givenKeys {
		value, err = dbCheck.Get(ro, k)
		defer value.Free()
		require.NoError(t, err)
		require.Equal(t, value.Data(), givenValue)
	}

}
