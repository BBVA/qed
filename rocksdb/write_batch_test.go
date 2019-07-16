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

func TestWriteBatch(t *testing.T) {

	db, path := newTestDB(t, "TestWriteBatch", nil)
	defer func() {
		db.Close()
		os.RemoveAll(path)
	}()

	var (
		key1   = []byte("key1")
		value1 = []byte("val1")
		key2   = []byte("key2")
	)
	wo := NewDefaultWriteOptions()
	require.NoError(t, db.Put(wo, key2, []byte("foo")))

	// create and fill the write batch
	wb := NewWriteBatch()
	defer wb.Destroy()
	wb.Put(key1, value1)
	wb.Delete(key2)
	require.Equal(t, wb.Count(), 2)

	// perform the batch write
	require.NoError(t, db.Write(wo, wb))

	// check changes
	ro := NewDefaultReadOptions()
	v1, err := db.Get(ro, key1)
	defer v1.Free()
	require.NoError(t, err)
	require.Equal(t, v1.Data(), value1)

	v2, err := db.Get(ro, key2)
	defer v2.Free()
	require.NoError(t, err)
	require.Nil(t, v2.Data())

}

func TestDeleteRange(t *testing.T) {

	db, path := newTestDB(t, "TestDeleteRange", nil)
	defer func() {
		db.Close()
		os.RemoveAll(path)
	}()

	wo := NewDefaultWriteOptions()
	defer wo.Destroy()
	ro := NewDefaultReadOptions()
	defer ro.Destroy()

	var (
		key1 = []byte("key1")
		key2 = []byte("key2")
		key3 = []byte("key3")
		key4 = []byte("key4")
		val1 = []byte("value")
		val2 = []byte("12345678")
		val3 = []byte("abcdefg")
		val4 = []byte("xyz")
	)

	require.NoError(t, db.Put(wo, key1, val1))
	require.NoError(t, db.Put(wo, key2, val2))
	require.NoError(t, db.Put(wo, key3, val3))
	require.NoError(t, db.Put(wo, key4, val4))

	actualVal1, err := db.GetBytes(ro, key1)
	require.NoError(t, err)
	require.Equal(t, actualVal1, val1)
	actualVal2, err := db.GetBytes(ro, key2)
	require.NoError(t, err)
	require.Equal(t, actualVal2, val2)
	actualVal3, err := db.GetBytes(ro, key3)
	require.NoError(t, err)
	require.Equal(t, actualVal3, val3)
	actualVal4, err := db.GetBytes(ro, key4)
	require.NoError(t, err)
	require.Equal(t, actualVal4, val4)

	batch := NewWriteBatch()
	defer batch.Destroy()
	batch.DeleteRange(key2, key4)
	_ = db.Write(wo, batch)

	actualVal1, err = db.GetBytes(ro, key1)
	require.NoError(t, err)
	require.Equal(t, actualVal1, val1)
	actualVal2, err = db.GetBytes(ro, key2)
	require.NoError(t, err)
	require.Nil(t, actualVal2)
	actualVal3, err = db.GetBytes(ro, key3)
	require.NoError(t, err)
	require.Nil(t, actualVal3)
	actualVal4, err = db.GetBytes(ro, key4)
	require.NoError(t, err)
	require.Equal(t, actualVal4, val4)

}
