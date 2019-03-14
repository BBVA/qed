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

// #include <rocksdb/c.h>
import "C"

// WriteBatch holds a collection of updates to apply atomically to a DB.
//
// The updates are applied in the order in which they are added
// to the WriteBatch.  For example, the value of "key" will be "v3"
// after the following batch is written:
//
//    batch.Put("key", "v1");
//    batch.Delete("key");
//    batch.Put("key", "v2");
//    batch.Put("key", "v3");
//
type WriteBatch struct {
	batch *C.rocksdb_writebatch_t
}

// NewWriteBatch create a WriteBatch object.
func NewWriteBatch() *WriteBatch {
	return &WriteBatch{batch: C.rocksdb_writebatch_create()}
}

// Put stores the mapping "key->value" in the database.
func (wb *WriteBatch) Put(key, value []byte) {
	cKey := bytesToChar(key)
	cValue := bytesToChar(value)
	C.rocksdb_writebatch_put(wb.batch, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)))
}

// Delete erases the mapping for "key" if it exists. Else, do nothing.
func (wb *WriteBatch) Delete(key []byte) {
	cKey := bytesToChar(key)
	C.rocksdb_writebatch_delete(wb.batch, cKey, C.size_t(len(key)))
}

// WriteBatch implementation of DeleteRange() // TODO

// Merge "value" with the existing value of "key" in the database.
// "key->merge(existing, value)"
func (wb *WriteBatch) Merge(key, value []byte) {
	cKey := bytesToChar(key)
	cValue := bytesToChar(value)
	C.rocksdb_writebatch_merge(wb.batch, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)))
}

// Clear all updates buffered in this batch.
func (wb *WriteBatch) Clear() {
	C.rocksdb_writebatch_clear(wb.batch)
}

// Count returns the number of updates in the batch.
func (wb *WriteBatch) Count() int {
	return int(C.rocksdb_writebatch_count(wb.batch))
}

// Destroy deallocates the WriteBatch object.
func (wb *WriteBatch) Destroy() {
	C.rocksdb_writebatch_destroy(wb.batch)
	wb.batch = nil
}
