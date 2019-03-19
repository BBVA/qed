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
// #include <stdlib.h>
import "C"
import (
	"errors"
	"unsafe"
)

// DB is a reusable handler to a RocksDB database on disk, created by OpenDB.
type DB struct {
	db   *C.rocksdb_t
	opts *Options
}

// OpenDB opens a database with the specified options.
func OpenDB(path string, opts *Options) (*DB, error) {
	var cErr *C.char
	cPath := C.CString(path)
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

// OpenDBForReadOnly opens a database with the specified options for readonly usage.
func OpenDBForReadOnly(path string, opts *Options, errorIfLogFileExist bool) (*DB, error) {
	var cErr *C.char
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	db := C.rocksdb_open_for_read_only(opts.opts, cPath, boolToUchar(errorIfLogFileExist), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}

	return &DB{
		db:   db,
		opts: opts,
	}, nil
}

// Close closes the database.
func (db *DB) Close() error {
	if db.db != nil {
		C.rocksdb_close(db.db)
		db.db = nil
	}
	db.opts.Destroy()
	return nil
}

// NewCheckpoint creates a new Checkpoint for this db.
func (db *DB) NewCheckpoint() (*Checkpoint, error) {
	var cErr *C.char
	cCheckpoint := C.rocksdb_checkpoint_object_create(db.db, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return NewNativeCheckpoint(cCheckpoint), nil
}

// Put writes data associated with a key to the database.
func (db *DB) Put(opts *WriteOptions, key, value []byte) error {
	cKey := bytesToChar(key)
	cValue := bytesToChar(value)
	var cErr *C.char
	C.rocksdb_put(db.db, opts.opts, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Get returns the data associated with the key from the database.
func (db *DB) Get(opts *ReadOptions, key []byte) (*Slice, error) {
	var cErr *C.char
	var cValueLen C.size_t
	cKey := bytesToChar(key)
	cValue := C.rocksdb_get(db.db, opts.opts, cKey, C.size_t(len(key)), &cValueLen, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return NewSlice(cValue, cValueLen), nil
}

// GetBytes is like Get but returns a copy of the data instead of a Slice.
func (db *DB) GetBytes(opts *ReadOptions, key []byte) ([]byte, error) {
	var cErr *C.char
	var cValueLen C.size_t
	cKey := bytesToChar(key)
	cValue := C.rocksdb_get(db.db, opts.opts, cKey, C.size_t(len(key)), &cValueLen, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	if cValue == nil {
		return nil, nil
	}
	defer C.free(unsafe.Pointer(cValue))
	return C.GoBytes(unsafe.Pointer(cValue), C.int(cValueLen)), nil
}

// Delete removes the data associated with the key from the database.
func (db *DB) Delete(opts *WriteOptions, key []byte) error {
	var cErr *C.char
	cKey := bytesToChar(key)
	C.rocksdb_delete(db.db, opts.opts, cKey, C.size_t(len(key)), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Write writes a WriteBatch to the database
func (db *DB) Write(opts *WriteOptions, batch *WriteBatch) error {
	var cErr *C.char
	C.rocksdb_write(db.db, opts.opts, batch.batch, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// NewIterator returns an Iterator over the the database that uses the
// ReadOptions given.
func (db *DB) NewIterator(opts *ReadOptions) *Iterator {
	cIter := C.rocksdb_create_iterator(db.db, opts.opts)
	return NewNativeIterator(unsafe.Pointer(cIter))
}

// Flush triggers a manuel flush for the database.
func (db *DB) Flush(opts *FlushOptions) error {
	var cErr *C.char
	C.rocksdb_flush(db.db, opts.opts, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}
