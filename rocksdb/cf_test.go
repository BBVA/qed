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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenDBColumnFamilies(t *testing.T) {

	dir, err := ioutil.TempDir("", "rocksdb-TestOpenDBColumnFamilies")
	require.NoError(t, err)

	givenNames := []string{"default", "other"}
	opts := NewDefaultOptions()
	opts.SetCreateIfMissingColumnFamilies(true)
	opts.SetCreateIfMissing(true)

	db, cfh, err := OpenDBColumnFamilies(
		dir, opts, givenNames, []*Options{opts, opts},
	)
	require.NoError(t, err)
	defer db.Close()

	require.Equal(t, len(cfh), 2)
	cfh[0].Destroy()
	cfh[1].Destroy()

	actualNames, err := ListColumnFamilies(dir, opts)
	require.NoError(t, err)
	require.ElementsMatch(t, actualNames, givenNames)

}

func TestColumnFamilyBatchPutGet(t *testing.T) {
	db, cfh, closeF := newTestDBCF(t, "TestColumnFamilyBatchPutGet")
	defer closeF()

	wo := NewDefaultWriteOptions()
	defer wo.Destroy()
	ro := NewDefaultReadOptions()
	defer ro.Destroy()

	key0 := []byte("hello0")
	value0 := []byte("world0")
	key1 := []byte("hello1")
	value1 := []byte("world1")

	batch0 := NewWriteBatch()
	defer batch0.Destroy()
	batch0.PutCF(cfh[0], key0, value0)
	require.NoError(t, db.Write(wo, batch0))
	actualValue0, err := db.GetCF(ro, cfh[0], key0)
	defer actualValue0.Free()
	require.NoError(t, err)
	require.Equal(t, actualValue0.Data(), value0)

	batch1 := NewWriteBatch()
	defer batch1.Destroy()
	batch1.PutCF(cfh[1], key1, value1)
	require.NoError(t, db.Write(wo, batch1))
	actualValue1, err := db.GetCF(ro, cfh[1], key1)
	defer actualValue1.Free()
	require.NoError(t, err)
	require.Equal(t, actualValue1.Data(), value1)

	// check the keys are not inserted in different CF
	actualValue, err := db.GetCF(ro, cfh[0], key1)
	require.NoError(t, err)
	require.Equal(t, actualValue.Size(), 0)
	actualValue, err = db.GetCF(ro, cfh[1], key0)
	require.NoError(t, err)
	require.Equal(t, actualValue.Size(), 0)

}

func TestColumnFamilyPutGetDelete(t *testing.T) {
	db, cfh, closeF := newTestDBCF(t, "TestColumnFamilyPutGetDelete")
	defer closeF()

	wo := NewDefaultWriteOptions()
	defer wo.Destroy()
	ro := NewDefaultReadOptions()
	defer ro.Destroy()

	key0 := []byte("hello0")
	value0 := []byte("world0")
	key1 := []byte("hello1")
	value1 := []byte("world1")

	require.NoError(t, db.PutCF(wo, cfh[0], key0, value0))
	actualValue0, err := db.GetCF(ro, cfh[0], key0)
	defer actualValue0.Free()
	require.NoError(t, err)
	require.Equal(t, actualValue0.Data(), value0)

	require.NoError(t, db.PutCF(wo, cfh[1], key1, value1))
	actualValue1, err := db.GetCF(ro, cfh[1], key1)
	defer actualValue1.Free()
	require.NoError(t, err)
	require.Equal(t, actualValue1.Data(), value1)

	actualValue, err := db.GetCF(ro, cfh[0], key1)
	require.NoError(t, err)
	require.Equal(t, actualValue.Size(), 0)
	actualValue, err = db.GetCF(ro, cfh[1], key0)
	require.NoError(t, err)
	require.Equal(t, actualValue.Size(), 0)

	require.NoError(t, db.DeleteCF(wo, cfh[0], key0))
	actualValue, err = db.GetCF(ro, cfh[0], key0)
	require.NoError(t, err)

}

func TestDeleteRangeCF(t *testing.T) {

	db, cfh, closeF := newTestDBCF(t, "TestColumnFamilyPutGetDelete")
	defer closeF()

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

	require.NoError(t, db.PutCF(wo, cfh[0], key1, val1))
	require.NoError(t, db.PutCF(wo, cfh[0], key2, val2))
	require.NoError(t, db.PutCF(wo, cfh[1], key3, val3))
	require.NoError(t, db.PutCF(wo, cfh[1], key4, val4))

	actualVal1, err := db.GetBytesCF(ro, cfh[0], key1)
	require.NoError(t, err)
	require.Equal(t, actualVal1, val1)
	actualVal2, err := db.GetBytesCF(ro, cfh[0], key2)
	require.NoError(t, err)
	require.Equal(t, actualVal2, val2)
	actualVal3, err := db.GetBytesCF(ro, cfh[1], key3)
	require.NoError(t, err)
	require.Equal(t, actualVal3, val3)
	actualVal4, err := db.GetBytesCF(ro, cfh[1], key4)
	require.NoError(t, err)
	require.Equal(t, actualVal4, val4)

	batch := NewWriteBatch()
	defer batch.Destroy()
	batch.DeleteRangeCF(cfh[0], key2, key4) // only keys from "defaul" cf
	_ = db.Write(wo, batch)

	actualVal1, err = db.GetBytesCF(ro, cfh[0], key1)
	require.NoError(t, err)
	require.Equal(t, actualVal1, val1)
	actualVal2, err = db.GetBytesCF(ro, cfh[0], key2)
	require.NoError(t, err)
	require.Nil(t, actualVal2) // <- the only one deleted
	actualVal3, err = db.GetBytesCF(ro, cfh[1], key3)
	require.NoError(t, err)
	require.Equal(t, actualVal3, val3)
	actualVal4, err = db.GetBytesCF(ro, cfh[1], key4)
	require.NoError(t, err)
	require.Equal(t, actualVal4, val4)

}
