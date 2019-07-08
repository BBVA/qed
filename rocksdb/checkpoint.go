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
	"errors"
	"unsafe"
)

// Checkpoint provides Checkpoint functionality.
// A checkpoint is an openable snapshot of a database at a point in time.
type Checkpoint struct {
	c *C.rocksdb_checkpoint_t
}

// NewNativeCheckpoint creates a new checkpoint.
func NewNativeCheckpoint(c *C.rocksdb_checkpoint_t) *Checkpoint {
	return &Checkpoint{c: c}
}

// CreateCheckpoint builds an openable snapshot of RocksDB on the same disk, which
// accepts an output directory on the same disk, and under the directory
// (1) hard-linked SST files pointing to existing live SST files
// SST files will be copied if output directory is on a different filesystem
// (2) a copied manifest files and other files
// The directory should not already exist and will be created by this API.
// The directory will be an absolute path
// logSizeForFlush: if the total log file size is equal or larger than
// this value, then a flush is triggered for all the column families. The
// default value is 0, which means flush is always triggered. If you move
// away from the default, the checkpoint may not contain up-to-date data
// if WAL writing is not always enabled.
// Flush will always trigger if it is 2PC.
func (cp *Checkpoint) CreateCheckpoint(checkpointDir string, logSizeForFlush uint64) error {
	var cErr *C.char
	cDir := C.CString(checkpointDir)
	defer C.free(unsafe.Pointer(cDir))

	C.rocksdb_checkpoint_create(cp.c, cDir, C.uint64_t(logSizeForFlush), &cErr)
	if cErr != nil {
		defer C.free(unsafe.Pointer(cErr))
		return errors.New(C.GoString(cErr))
	}
	return nil
}

// Destroy deallocates the Checkpoint object.
func (cp *Checkpoint) Destroy() {
	C.rocksdb_checkpoint_object_destroy(cp.c)
	cp.c = nil
}
