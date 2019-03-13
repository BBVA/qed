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

// WriteOptions represent all options available when writing to a database.
type WriteOptions struct {
	opts *C.rocksdb_writeoptions_t
}

// NewDefaultWriteOptions creates a default WriteOptions object.
func NewDefaultWriteOptions() *WriteOptions {
	return &WriteOptions{C.rocksdb_writeoptions_create()}
}

// SetDisableWAL sets whether WAL should be active or not.
// If true, writes will not first go to the write ahead log,
// and the write may got lost after a crash.
// Default: false
func (o *WriteOptions) SetDisableWAL(value bool) {
	C.rocksdb_writeoptions_disable_WAL(o.opts, C.int(btoi(value)))
}

// Destroy deallocates the WriteOptions object.
func (o *WriteOptions) Destroy() {
	C.rocksdb_writeoptions_destroy(o.opts)
	o.opts = nil
}