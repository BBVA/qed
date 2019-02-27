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
// #cgo LDFLAGS: -lrocksdb
// #include "../c-deps/rocksdb/include/rocksdb/c.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"unsafe"
)

// DB is a reusable handler to a RocksDB database on disk, created by Open.
type DB struct {
	db   *C.rocksdb_t
	opts *Options
}

func Open(path string, opts *Options) (*DB, error) {
	var cErr *C.char
	var cPath = C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	db := C.rocksdb_open(opts.opts, cPath, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}

	return &DB{
		db:   db,
		opts: opts,
	}, nil
}

func (db *DB) Close() error {
	if db.db != nil {
		C.rocksdb_close(db.db)
		db.db = nil
	}
	db.opts.Destroy()
	return nil
}
