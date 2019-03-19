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

type ReadOptions struct {
	opts *C.rocksdb_readoptions_t
}

// NewDefaultReadOptions creates a default ReadOptions object.
func NewDefaultReadOptions() *ReadOptions {
	return &ReadOptions{opts: C.rocksdb_readoptions_create()}
}

// SetFillCache specify whether the "data block"/"index block"/"filter block"
// read for this iteration should be cached in memory?
// Callers may wish to set this field to false for bulk scans.
// Default: true
func (o *ReadOptions) SetFillCache(value bool) {
	C.rocksdb_readoptions_set_fill_cache(o.opts, boolToUchar(value))
}

// Destroy deallocates the ReadOptions object.
func (o *ReadOptions) Destroy() {
	C.rocksdb_readoptions_destroy(o.opts)
	o.opts = nil
}
