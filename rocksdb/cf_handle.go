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
// #include <rocksdb/c.h>
import "C"
import (
	"reflect"
	"unsafe"
)

// ColumnFamilyHandle represents a handle to a ColumnFamily.
type ColumnFamilyHandle struct {
	c *C.rocksdb_column_family_handle_t
}

// NewColumnFamilyHandle creates a ColumnFamilyHandle object.
func NewColumnFamilyHandle(c *C.rocksdb_column_family_handle_t) *ColumnFamilyHandle {
	return &ColumnFamilyHandle{c}
}

// UnsafeGetCFHandler returns the underlying c column family handle.
func (h *ColumnFamilyHandle) UnsafeGetCFHandler() unsafe.Pointer {
	return unsafe.Pointer(h.c)
}

// Destroy calls the destructor of the underlying column family handle.
func (h *ColumnFamilyHandle) Destroy() {
	C.rocksdb_column_family_handle_destroy(h.c)
}

type ColumnFamilyHandles []*ColumnFamilyHandle

func (cfs ColumnFamilyHandles) toCSlice() columnFamilySlice {
	cCFs := make(columnFamilySlice, len(cfs))
	for i, cf := range cfs {
		cCFs[i] = cf.c
	}
	return cCFs
}

type columnFamilySlice []*C.rocksdb_column_family_handle_t

func (s columnFamilySlice) c() **C.rocksdb_column_family_handle_t {
	sH := (*reflect.SliceHeader)(unsafe.Pointer(&s))
	return (**C.rocksdb_column_family_handle_t)(unsafe.Pointer(sH.Data))
}
