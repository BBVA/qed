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
package rocksdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteBatch(t *testing.T) {

	db := newTestDB(t, "TestWriteBatch", nil)
	defer db.Close()

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
