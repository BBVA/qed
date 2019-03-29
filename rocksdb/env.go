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
import "C"

// Env is an environment used to route through all file operations and
// some system calls.
type Env struct {
	c *C.rocksdb_env_t
}

// NewDefaultEnv creates a default environment.
func NewDefaultEnv() *Env {
	return &Env{C.rocksdb_create_default_env()}
}

// SetBackgroundThreads sets the number of background worker threads
// of a specific thread pool for this environment.
// 'LOW' is the default pool.
// Default: 1
func (e *Env) SetBackgroundThreads(n int) {
	C.rocksdb_env_set_background_threads(e.c, C.int(n))
}

// SetHighPriorityBackgroundThreads sets the size of the high priority
// thread pool that can be used to prevent compactions from stalling
// memtable flushes.
func (e *Env) SetHighPriorityBackgroundThreads(n int) {
	C.rocksdb_env_set_high_priority_background_threads(e.c, C.int(n))
}

// Destroy deallocates the Env object.
func (e *Env) Destroy() {
	C.rocksdb_env_destroy(e.c)
	e.c = nil
}
