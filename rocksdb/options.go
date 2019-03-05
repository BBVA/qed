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
// #include <rocksdb/c.h>
import "C"
import "unsafe"

// CompressionType specifies the block compression.
// DB contents are stored in a set of blocks, each of which holds a
// sequence of key,value pairs. Each block may be compressed before
// being stored in a file. The following enum describes which
// compression method (if any) is used to compress a block.
type CompressionType uint

// Compression types
const (
	NoCompression     = CompressionType(C.rocksdb_no_compression)
	SnappyCompression = CompressionType(C.rocksdb_snappy_compression)
)

// Options represent all of the available options when opening a database with Open.
type Options struct {
	opts *C.rocksdb_options_t

	// Hold references for GC.
	bbto *BlockBasedTableOptions
}

// NewDefaultOptions creates the default Options.
func NewDefaultOptions() *Options {
	return &Options{opts: C.rocksdb_options_create()}
}

// SetCreateIfMissing specifies whether the database
// should be created if it is missing.
// Default: false
func (o *Options) SetCreateIfMissing(value bool) {
	C.rocksdb_options_set_create_if_missing(o.opts, boolToUchar(value))
}

// IncreaseParallelism sets the level of parallelism.
//
// By default, RocksDB uses only one background thread for flush and
// compaction. Calling this function will set it up such that total of
// `totalThreads` is used. Good value for `totalThreads` is the number of
// cores. You almost definitely want to call this function if your system is
// bottlenecked by RocksDB.
func (o *Options) IncreaseParallelism(totalThreads int) {
	C.rocksdb_options_increase_parallelism(o.opts, C.int(totalThreads))
}

// SetMaxWriteBufferNumber sets the maximum number of write buffers (memtables)
// that are built up in memory.
//
// The default is 2, so that when 1 write buffer is being flushed to
// storage, new writes can continue to the other write buffer.
// Default: 2
func (o *Options) SetMaxWriteBufferNumber(value int) {
	C.rocksdb_options_set_max_write_buffer_number(o.opts, C.int(value))
}

// SetMinWriteBufferNumberToMerge sets the minimum number of write buffers
// that will be merged together before writing to storage.
//
// If set to 1, then all write buffers are flushed to L0 as individual files
// and this increases read amplification because a get request has to check
// in all of these files. Also, an in-memory merge may result in writing lesser
// data to storage if there are duplicate records in each of these
// individual write buffers.
// Default: 1
func (o *Options) SetMinWriteBufferNumberToMerge(value int) {
	C.rocksdb_options_set_min_write_buffer_number_to_merge(o.opts, C.int(value))
}

// SetBlockBasedTableFactory sets the block based table factory.
func (o *Options) SetBlockBasedTableFactory(value *BlockBasedTableOptions) {
	o.bbto = value
	C.rocksdb_options_set_block_based_table_factory(o.opts, value.opts)
}

// SetDBLogDir specifies the absolute info LOG dir.
//
// If it is empty, the log files will be in the same dir as data.
// If it is non empty, the log files will be in the specified dir,
// and the db data dir's absolute path will be used as the log file
// name's prefix.
// Default: empty
func (o *Options) SetDBLogDir(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.rocksdb_options_set_db_log_dir(o.opts, cValue)
}

// SetWalDir specifies the absolute dir path for write-ahead logs (WAL).
//
// If it is empty, the log files will be in the same dir as data.
// If it is non empty, the log files will be in the specified dir,
// When destroying the db, all log files and the dir itopts is deleted.
// Default: empty
func (o *Options) SetWalDir(value string) {
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))
	C.rocksdb_options_set_wal_dir(o.opts, cValue)
}

// Destroy deallocates the Options object.
func (o *Options) Destroy() {
	C.rocksdb_options_destroy(o.opts)
	if o.bbto != nil {
		o.bbto.Destroy()
	}
	o.opts = nil
	o.bbto = nil
}
