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
import "C"
import "unsafe"

// Slice is a simple structure that contains a length and a
// pointer to an external byte array. It is used as a wrapper
// for non-copy values.
// Be careful when using Slices since it is up to the caller
// to ensure that the external byte array into which the Slice
// points remains live while the Slice is in use.
type Slice struct {
	data  *C.char
	size  C.size_t
	freed bool
}

// NewSlice returns a slice with the given data.
func NewSlice(data *C.char, size C.size_t) *Slice {
	return &Slice{
		data:  data,
		size:  size,
		freed: false,
	}
}

// Data returns the data of the slice.
func (s *Slice) Data() []byte {
	return charToBytes(s.data, s.size)
}

// Size returns the size of the data.
func (s *Slice) Size() int {
	return int(s.size)
}

// Free frees the slice data.
func (s *Slice) Free() {
	if !s.freed {
		C.free(unsafe.Pointer(s.data))
		s.freed = true
	}
}
