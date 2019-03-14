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

// #include <stdlib.h>
// #include "rocksdb/c.h"
import "C"
import (
	"bytes"
	"errors"
	"unsafe"
)

// Iterator provides a way to seek to specific keys and iterate through
// the keyspace from that point, as well as access the values of those keys.
//
// For example:
//
//      it := db.NewIterator(readOpts)
//      defer it.Close()
//
//      it.Seek([]byte("foo"))
//		for ; it.Valid(); it.Next() {
//          fmt.Printf("Key: %v Value: %v\n", it.Key().Data(), it.Value().Data())
// 		}
//
//      if err := it.Err(); err != nil {
//          return err
//      }
//
type Iterator struct {
	it *C.rocksdb_iterator_t
}

// NewNativeIterator creates a Iterator object.
func NewNativeIterator(c unsafe.Pointer) *Iterator {
	return &Iterator{it: (*C.rocksdb_iterator_t)(c)}
}

// Valid returns false only when an Iterator has iterated past either the
// first or the last key in the database. An iterator is either positioned
// at a key/value pair, or not valid.
func (iter *Iterator) Valid() bool {
	return C.rocksdb_iter_valid(iter.it) != 0
}

// ValidForPrefix returns false only when an Iterator has iterated past the
// first or the last key in the database or the specified prefix.
func (iter *Iterator) ValidForPrefix(prefix []byte) bool {
	if C.rocksdb_iter_valid(iter.it) == 0 {
		return false
	}
	keySlice := iter.Key()
	result := bytes.HasPrefix(keySlice.Data(), prefix)
	keySlice.Free()
	return result
}

// Key returns the key the iterator currently holds.
// The underlying storage for the returned slice is valid
// only until the next modification of the iterator.
// REQUIRES: Valid()
func (iter *Iterator) Key() *Slice {
	var cLen C.size_t
	cKey := C.rocksdb_iter_key(iter.it, &cLen)
	if cKey == nil {
		return nil
	}
	return &Slice{cKey, cLen, true}
}

// Value returns the value in the database the iterator currently holds.
// The underlying storage for the returned slice is valid
// only until the next modification of the iterator.
// REQUIRES: Valid()
func (iter *Iterator) Value() *Slice {
	var cLen C.size_t
	cVal := C.rocksdb_iter_value(iter.it, &cLen)
	if cVal == nil {
		return nil
	}
	return &Slice{cVal, cLen, true}
}

// Next moves the iterator to the next sequential key in the database.
// After this call, Valid() is true if the iterator was not positioned
// at the last entry in the source.
// REQUIRES: Valid()
func (iter *Iterator) Next() {
	C.rocksdb_iter_next(iter.it)
}

// Prev moves the iterator to the previous sequential key in the database.
// After this call, Valid() is true if the iterator was not positioned at
// the first entry in source.
// REQUIRES: Valid()
func (iter *Iterator) Prev() {
	C.rocksdb_iter_prev(iter.it)
}

// SeekToFirst moves the iterator to the first key in the database.
// The iterator is Valid() after this call if the source is not empty.
func (iter *Iterator) SeekToFirst() {
	C.rocksdb_iter_seek_to_first(iter.it)
}

// SeekToLast moves the iterator to the last key in the database.
// The iterator is Valid() after this call if the source is not empty.
func (iter *Iterator) SeekToLast() {
	C.rocksdb_iter_seek_to_last(iter.it)
}

// Seek moves the iterator to the position greater than or equal to the key.
// The iterator is Valid() after this call if the source contains
// an entry that comes at or past target.
// All Seek*() methods clear any error that the iterator had prior to
// the call; after the seek, Error() indicates only the error (if any) that
// happened during the seek, not any past errors.
func (iter *Iterator) Seek(key []byte) {
	cKey := bytesToChar(key)
	C.rocksdb_iter_seek(iter.it, cKey, C.size_t(len(key)))
}

// SeekForPrev moves the iterator to the last key that less than or equal
// to the target key, in contrast with Seek.
// The iterator is Valid() after this call if the source contains
// an entry that comes at or before target.
func (iter *Iterator) SeekForPrev(key []byte) {
	cKey := bytesToChar(key)
	C.rocksdb_iter_seek_for_prev(iter.it, cKey, C.size_t(len(key)))
}

// Err returns nil if no errors happened during iteration, or the actual
// error otherwise.
func (iter *Iterator) Err() error {
	var cErr *C.char
	C.rocksdb_iter_get_error(iter.it, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Close closes the iterator.
func (iter *Iterator) Close() {
	C.rocksdb_iter_destroy(iter.it)
	iter.it = nil
}
