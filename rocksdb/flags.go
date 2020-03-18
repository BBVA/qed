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

// #cgo CFLAGS: -I${SRCDIR}/../c-deps/rocksdb/include
// #cgo CXXFLAGS: -std=c++11 -O3 -I${SRCDIR}/../c-deps/rocksdb/include
// #cgo LDFLAGS: -L${SRCDIR}/../c-deps/libs
// #cgo LDFLAGS: -lrocksdb
// #cgo !dragonfly LDFLAGS: -ljemalloc
// #cgo LDFLAGS: -lsnappy
// #cgo LDFLAGS: -lstdc++
// #cgo LDFLAGS: -ldl
// #cgo LDFLAGS: -lpthread
// #cgo LDFLAGS: -lm
// #cgo darwin LDFLAGS: -Wl,-undefined -Wl,dynamic_lookup
// #cgo linux LDFLAGS: -Wl,-unresolved_symbols=ignore-all -lrt
import "C"
