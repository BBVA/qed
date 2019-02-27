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

// #cgo LDFLAGS: -L../c-deps/rocksdb/
// #cgo LDFLAGS: -l:librocksdb.a -l:libstdc++.a -l:libz.a -l:libbz2.a -l:libsnappy.a -l:libzstd.a -lm
// #include "../c-deps/rocksdb/include/rocksdb/c.h"
import "C"

type CompressionType uint

// Compression types
const (
	NoCompression     = CompressionType(C.rocksdb_no_compression)
	SnappyCompression = CompressionType(C.rocksdb_snappy_compression)
)

// Options represent all of the available options when opening a database with Open.
type Options struct {
	opts *C.rocksdb_options_t
}

func NewDefaultOptions() *Options {
	return NewNativeOptions(C.rocksdb_options_create())
}

func NewNativeOptions(opts *C.rocksdb_options_t) *Options {
	return &Options{opts: opts}
}

// SetCreateIfMissing specifies whether the database
// should be created if it is missing.
// Default: false
func (o *Options) SetCreateIfMissing(value bool) {
	C.rocksdb_options_set_create_if_missing(o.opts, boolToUchar(value))
}

func (o *Options) Destroy() {
	C.rocksdb_options_destroy(o.opts)
	o.opts = nil
}
