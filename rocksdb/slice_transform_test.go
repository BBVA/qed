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

func TestSliceTransform(t *testing.T) {

	db, path := newTestDB(t, "TestSliceTransform", func(opts *Options) {
		opts.SetPrefixExtractor(&testSliceTransform{})
	})
	defer func() {
		db.Close()
		os.RemoveAll(path)
	}()

	wo := NewDefaultWriteOptions()
	require.NoError(t, db.Put(wo, []byte("foo1"), []byte("foo")))
	require.NoError(t, db.Put(wo, []byte("foo2"), []byte("foo")))
	require.NoError(t, db.Put(wo, []byte("bar1"), []byte("bar")))

	it := db.NewIterator(NewDefaultReadOptions())
	defer it.Close()

	prefix := []byte("foo")
	numFound := 0
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		numFound++
	}
	require.Nil(t, it.Err())
	require.Equal(t, numFound, 2)

}

func TestFixedPrefixTransform(t *testing.T) {
	db, path := newTestDB(t, "TestFixedPrefixTransform", func(opts *Options) {
		opts.SetPrefixExtractor(NewFixedPrefixTransform(3))
	})
	defer func() {
		db.Close()
		os.RemoveAll(path)
	}()

	wo := NewDefaultWriteOptions()
	require.NoError(t, db.Put(wo, []byte("foo1"), []byte("foo")))
	require.NoError(t, db.Put(wo, []byte("foo2"), []byte("foo")))
	require.NoError(t, db.Put(wo, []byte("bar1"), []byte("bar")))

	it := db.NewIterator(NewDefaultReadOptions())
	defer it.Close()

	prefix := []byte("foo")
	numFound := 0
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		numFound++
	}
	require.Nil(t, it.Err())
	require.Equal(t, numFound, 2)
}

type testSliceTransform struct {
	initiated bool
}

func (st *testSliceTransform) Name() string {
	return "rocksdb.test"
}

func (st *testSliceTransform) Transform(key []byte) []byte {
	return key[0:3]
}

func (st *testSliceTransform) InDomain(key []byte) bool {
	return len(key) >= 3
}

func (st *testSliceTransform) InRange(key []byte) bool {
	return len(key) == 3
}
