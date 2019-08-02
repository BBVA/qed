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

// #include <stdlib.h>
// #include "rocksdb/c.h"
import "C"
import (
	"errors"
	"unsafe"
)

// WALIterator is used to iterate over the transactions
// in a db. One run on the iterator is continuous, i.e. the
// iterator will stop at the beginning of any gap in sequences.
type WALIterator struct {
	c *C.rocksdb_wal_iterator_t
}

// NewNativeWALIterator creates a WALIterator object.
func NewNativeWALIterator(c unsafe.Pointer) *WALIterator {
	return &WALIterator{c: (*C.rocksdb_wal_iterator_t)(c)}
}

// Valid returns false only when an Iterator has iterated past either the
// first or the last key in the database. An iterator is either positioned
// at a key/value pair, or not valid.
func (iter *WALIterator) Valid() bool {
	return C.rocksdb_wal_iter_valid(iter.c) != 0
}

// Next moves the iterator to the next sequential key in the database.
// After this call, Valid() is true if the iterator was not positioned
// at the last entry in the source.
// REQUIRES: Valid()
func (iter *WALIterator) Next() {
	C.rocksdb_wal_iter_next(iter.c)
}

// Status returns an error if something has gone wrong or else
// it returns nil.
func (iter *WALIterator) Status() error {
	var cErr *C.char
	C.rocksdb_wal_iter_status(iter.c, &cErr)
	if cErr != nil {
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// GetBatch returns, if valid, the current write_batch and the sequence
// number of the earliest transaction contained in the batch.
// ONLY use if Valid() is true and status() is OK.
func (iter *WALIterator) GetBatch() (*WriteBatch, uint64) {
	var cSeqNum C.uint64_t
	cBatch := C.rocksdb_wal_iter_get_batch(iter.c, &cSeqNum)
	return NewNativeWriteBatch(cBatch), uint64(cSeqNum)
}

// Close closes the iterator.
func (iter *WALIterator) Close() {
	C.rocksdb_wal_iter_destroy(iter.c)
	iter.c = nil
}
