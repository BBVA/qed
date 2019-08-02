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

// #include "rocksdb/c.h"
// #include	"extended.h"
// #include <stdlib.h>
import "C"
import (
	"errors"
	"unsafe"
)

// DB is a reusable handler to a RocksDB database on disk, created by OpenDB.
type DB struct {
	c    *C.rocksdb_t
	opts *Options
}

// OpenDB opens a database with the specified options.
func OpenDB(path string, opts *Options) (*DB, error) {
	var cErr *C.char
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	db := C.rocksdb_open(opts.c, cPath, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}

	return &DB{
		c:    db,
		opts: opts,
	}, nil
}

// OpenDBForReadOnly opens a database with the specified options for read-only usage.
func OpenDBForReadOnly(path string, opts *Options, errorIfLogFileExist bool) (*DB, error) {
	var cErr *C.char
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	db := C.rocksdb_open_for_read_only(opts.c, cPath, boolToUchar(errorIfLogFileExist), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}

	return &DB{
		c:    db,
		opts: opts,
	}, nil
}

// OpenDBColumnFamilies opens a database with the specified column families.
func OpenDBColumnFamilies(
	path string,
	opts *Options,
	cfNames []string,
	cfOpts []*Options,
) (*DB, ColumnFamilyHandles, error) {

	numColumnFamilies := len(cfNames)
	if numColumnFamilies != len(cfOpts) {
		return nil, nil, errors.New("must provide the same number of column family names and options")
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cNames := make([]*C.char, numColumnFamilies)
	for i, s := range cfNames {
		cNames[i] = C.CString(s)
	}
	defer func() {
		for _, s := range cNames {
			C.free(unsafe.Pointer(s))
		}
	}()

	cOpts := make([]*C.rocksdb_options_t, numColumnFamilies)
	for i, o := range cfOpts {
		cOpts[i] = o.c
	}

	cHandles := make([]*C.rocksdb_column_family_handle_t, numColumnFamilies)

	var cErr *C.char
	db := C.rocksdb_open_column_families(
		opts.c,
		cPath,
		C.int(numColumnFamilies),
		&cNames[0],
		&cOpts[0],
		&cHandles[0],
		&cErr,
	)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, nil, errors.New(C.GoString(cErr))
	}

	cfHandles := make([]*ColumnFamilyHandle, numColumnFamilies)
	for i, c := range cHandles {
		cfHandles[i] = NewColumnFamilyHandle(c)
	}

	return &DB{
		c:    db,
		opts: opts,
	}, cfHandles, nil
}

// OpenDBForReadOnlyColumnFamilies opens a database with the specified column
// families in read-only mode.
func OpenDBForReadOnlyColumnFamilies(
	path string,
	opts *Options,
	cfNames []string,
	cfOpts []*Options,
	errorIfLogFileExist bool,
) (*DB, ColumnFamilyHandles, error) {

	numColumnFamilies := len(cfNames)
	if numColumnFamilies != len(cfOpts) {
		return nil, nil, errors.New("must provide the same number of column family names and options")
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cNames := make([]*C.char, numColumnFamilies)
	for i, s := range cfNames {
		cNames[i] = C.CString(s)
	}
	defer func() {
		for _, s := range cNames {
			C.free(unsafe.Pointer(s))
		}
	}()

	cOpts := make([]*C.rocksdb_options_t, numColumnFamilies)
	for i, o := range cfOpts {
		cOpts[i] = o.c
	}

	cHandles := make([]*C.rocksdb_column_family_handle_t, numColumnFamilies)

	var cErr *C.char
	db := C.rocksdb_open_for_read_only_column_families(
		opts.c,
		cPath,
		C.int(numColumnFamilies),
		&cNames[0],
		&cOpts[0],
		&cHandles[0],
		boolToUchar(errorIfLogFileExist),
		&cErr,
	)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, nil, errors.New(C.GoString(cErr))
	}

	cfHandles := make([]*ColumnFamilyHandle, numColumnFamilies)
	for i, c := range cHandles {
		cfHandles[i] = NewColumnFamilyHandle(c)
	}

	return &DB{
		c:    db,
		opts: opts,
	}, cfHandles, nil
}

// ListColumnFamilies lists the names of the column families in the DB.
func ListColumnFamilies(path string, opts *Options) ([]string, error) {
	var cErr *C.char
	var cLen C.size_t
	var cPath = C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	cNames := C.rocksdb_list_column_families(opts.c, cPath, &cLen, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}

	namesLen := int(cLen)
	names := make([]string, namesLen)
	cNamesArr := (*[1 << 30]*C.char)(unsafe.Pointer(cNames))[:namesLen:namesLen]
	for i, n := range cNamesArr {
		names[i] = C.GoString(n)
	}

	C.rocksdb_list_column_families_destroy(cNames, cLen)
	return names, nil
}

// Close closes the database.
func (db *DB) Close() error {
	if db.c != nil {
		C.rocksdb_close(db.c)
		db.c = nil
	}
	return nil
}

// NewCheckpoint creates a new Checkpoint for this db.
func (db *DB) NewCheckpoint() (*Checkpoint, error) {
	var cErr *C.char
	cCheckpoint := C.rocksdb_checkpoint_object_create(db.c, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return NewNativeCheckpoint(cCheckpoint), nil
}

// Put writes data associated with a key to the database.
func (db *DB) Put(wo *WriteOptions, key, value []byte) error {
	cKey := bytesToChar(key)
	cValue := bytesToChar(value)
	var cErr *C.char
	C.rocksdb_put(db.c, wo.c, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// PutCF writes data associated with a key to the database and a column family.
func (db *DB) PutCF(wo *WriteOptions, cf *ColumnFamilyHandle, key, value []byte) error {
	cKey := bytesToChar(key)
	cValue := bytesToChar(value)
	var cErr *C.char
	C.rocksdb_put_cf(db.c, wo.c, cf.c, cKey, C.size_t(len(key)), cValue, C.size_t(len(value)), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Get returns the data associated with the key from the database.
func (db *DB) Get(ro *ReadOptions, key []byte) (*Slice, error) {
	var cErr *C.char
	var cValueLen C.size_t
	cKey := bytesToChar(key)
	cValue := C.rocksdb_get(db.c, ro.c, cKey, C.size_t(len(key)), &cValueLen, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return NewSlice(cValue, cValueLen), nil
}

// GetBytes is like Get but returns a copy of the data instead of a Slice.
func (db *DB) GetBytes(ro *ReadOptions, key []byte) ([]byte, error) {
	var cErr *C.char
	var cValueLen C.size_t
	cKey := bytesToChar(key)
	cValue := C.rocksdb_get(db.c, ro.c, cKey, C.size_t(len(key)), &cValueLen, &cErr)
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

// GetCF returns the data associated with the key from the database
// and column family.
func (db *DB) GetCF(ro *ReadOptions, cf *ColumnFamilyHandle, key []byte) (*Slice, error) {
	var cErr *C.char
	var cValueLen C.size_t
	cKey := bytesToChar(key)
	cValue := C.rocksdb_get_cf(db.c, ro.c, cf.c, cKey, C.size_t(len(key)), &cValueLen, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return NewSlice(cValue, cValueLen), nil
}

// GetBytesCF is like GetCF but returns a copy of the data instead of a Slice.
func (db *DB) GetBytesCF(ro *ReadOptions, cf *ColumnFamilyHandle, key []byte) ([]byte, error) {
	var cErr *C.char
	var cValueLen C.size_t
	cKey := bytesToChar(key)
	cValue := C.rocksdb_get_cf(db.c, ro.c, cf.c, cKey, C.size_t(len(key)), &cValueLen, &cErr)
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
func (db *DB) Delete(wo *WriteOptions, key []byte) error {
	var cErr *C.char
	cKey := bytesToChar(key)
	C.rocksdb_delete(db.c, wo.c, cKey, C.size_t(len(key)), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// DeleteCF removes the data associated with the key from the database and column family.
func (db *DB) DeleteCF(wo *WriteOptions, cf *ColumnFamilyHandle, key []byte) error {
	var cErr *C.char
	cKey := bytesToChar(key)
	C.rocksdb_delete_cf(db.c, wo.c, cf.c, cKey, C.size_t(len(key)), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Write writes a WriteBatch to the database
func (db *DB) Write(wo *WriteOptions, batch *WriteBatch) error {
	var cErr *C.char
	C.rocksdb_write(db.c, wo.c, batch.c, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// NewIterator returns an Iterator over the the database that uses the
// ReadOptions given.
func (db *DB) NewIterator(ro *ReadOptions) *Iterator {
	cIter := C.rocksdb_create_iterator(db.c, ro.c)
	return NewNativeIterator(unsafe.Pointer(cIter))
}

// NewIteratorCF returns an Iterator over the the database and column family
// that uses the ReadOptions given.
func (db *DB) NewIteratorCF(ro *ReadOptions, cf *ColumnFamilyHandle) *Iterator {
	cIter := C.rocksdb_create_iterator_cf(db.c, ro.c, cf.c)
	return NewNativeIterator(unsafe.Pointer(cIter))
}

// Flush triggers a manual flush for the database.
func (db *DB) Flush(fo *FlushOptions) error {
	var cErr *C.char
	C.rocksdb_flush(db.c, fo.c, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// GetProperty returns the value of a database property.
func (db *DB) GetProperty(propName string) string {
	cProp := C.CString(propName)
	defer C.free(unsafe.Pointer(cProp))
	cValue := C.rocksdb_property_value(db.c, cProp)
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue)
}

// GetPropertyCF returns the value of a database property.
func (db *DB) GetPropertyCF(propName string, cf *ColumnFamilyHandle) string {
	cProp := C.CString(propName)
	defer C.free(unsafe.Pointer(cProp))
	cValue := C.rocksdb_property_value_cf(db.c, cf.c, cProp)
	defer C.free(unsafe.Pointer(cValue))
	return C.GoString(cValue)
}

// GetUint64Property returns the value of a database property.
func (db *DB) GetUint64Property(propName string) uint64 {
	cProp := C.CString(propName)
	defer C.free(unsafe.Pointer(cProp))
	var cValue C.uint64_t
	C.rocksdb_property_int(db.c, cProp, &cValue)
	return uint64(cValue)
}

// GetUint64PropertyCF returns the value of a database property.
func (db *DB) GetUint64PropertyCF(propName string, cf *ColumnFamilyHandle) uint64 {
	cProp := C.CString(propName)
	defer C.free(unsafe.Pointer(cProp))
	var cValue C.uint64_t
	C.rocksdb_property_int_cf(db.c, cf.c, cProp, &cValue)
	return uint64(cValue)
}

// GetLatestSequenceNumber returns the sequence number of the most
// recent transaction.
func (db *DB) GetLatestSequenceNumber() uint64 {
	var cValue C.uint64_t
	cValue = C.rocksdb_get_latest_sequence_number(db.c)
	return uint64(cValue)
}

// GetUpdatesSince sets iter to an iterator that is positioned at a
// write-batch containing seq_number. If the sequence number is non existent,
// it returns an iterator at the first available seq_no after the requested seq_no.
// Returns an error if iterator is not valid.
// Must set WAL_ttl_seconds or WAL_size_limit_MB to large values to
// use this api, else the WAL files will get cleared aggressively and the
// iterator might keep getting invalid before an update is read.
func (db *DB) GetUpdatesSince(seqNum uint64) (*WALIterator, error) {
	var cErr *C.char
	var cOpts *C.rocksdb_wal_readoptions_t
	cIter := C.rocksdb_get_updates_since(db.c, C.uint64_t(seqNum), cOpts, &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return nil, errors.New(C.GoString(cErr))
	}
	return NewNativeWALIterator(unsafe.Pointer(cIter)), nil
}
