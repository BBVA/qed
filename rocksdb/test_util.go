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

func newTestDB(t require.TestingT, name string, applyOpts func(opts *Options)) (*DB, string) {
	path, err := ioutil.TempDir("/var/tmp", "rocksdb-"+name)
	require.NoError(t, err)

	opts := NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	if applyOpts != nil {
		applyOpts(opts)
	}

	db, err := OpenDB(path, opts)
	require.NoError(t, err)

	return db, path
}

func newTestDBfromPath(t require.TestingT, path string, applyOpts func(opts *Options)) *DB {
	opts := NewDefaultOptions()
	opts.SetCreateIfMissing(true)
	if applyOpts != nil {
		applyOpts(opts)
	}

	db, err := OpenDB(path, opts)
	require.NoError(t, err)

	return db
}

func newTestDBCF(t *testing.T, name string) (db *DB, cfh []*ColumnFamilyHandle, cleanup func()) {
	path, err := ioutil.TempDir("", "rocksdb-"+name)
	require.NoError(t, err)

	givenNames := []string{"default", "other"}
	opts := NewDefaultOptions()
	opts.SetCreateIfMissingColumnFamilies(true)
	opts.SetCreateIfMissing(true)

	db, cfh, err = OpenDBColumnFamilies(path, opts, givenNames, []*Options{opts, opts})
	require.NoError(t, err)

	cleanup = func() {
		for _, cf := range cfh {
			cf.Destroy()
		}
		db.Close()
	}
	return db, cfh, cleanup
}
