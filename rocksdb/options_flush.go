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

// #include <rocksdb/c.h>
import "C"

// FlushOptions represent all of the available options when manual flushing the
// database.
type FlushOptions struct {
	opts *C.rocksdb_flushoptions_t
}

// NewDefaultFlushOptions creates a default FlushOptions object.
func NewDefaultFlushOptions() *FlushOptions {
	return &FlushOptions{C.rocksdb_flushoptions_create()}
}

// SetWait specify if the flush will wait until the flush is done.
// Default: true
func (o *FlushOptions) SetWait(value bool) {
	C.rocksdb_flushoptions_set_wait(o.opts, boolToUchar(value))
}

// Destroy deallocates the FlushOptions object.
func (o *FlushOptions) Destroy() {
	C.rocksdb_flushoptions_destroy(o.opts)
	o.opts = nil
}
