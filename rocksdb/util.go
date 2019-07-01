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
import (
	"reflect"
	"unsafe"
)

// btoi converts a bool value to int.
func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}

// boolToUchar converts a bool value to C.uchar.
func boolToUchar(b bool) C.uchar {
	if b {
		return C.uchar(1)
	}
	return C.uchar(0)
}

// bytesToChar converts a byte slice to *C.char.
func bytesToChar(b []byte) *C.char {
	var c *C.char
	if len(b) > 0 {
		c = (*C.char)(unsafe.Pointer(&b[0]))
	}
	return c
}

// Go []byte to C string
// The C string is allocated in the C heap using malloc.
func cByteSlice(b []byte) *C.char {
	var c *C.char
	if len(b) > 0 {
		cData := C.malloc(C.size_t(len(b)))
		copy((*[1 << 24]byte)(cData)[0:len(b)], b)
		c = (*C.char)(cData)
	}
	return c
}

// charToBytes converts a *C.char to a byte slice.
func charToBytes(data *C.char, len C.size_t) []byte {
	var value []byte
	header := (*reflect.SliceHeader)(unsafe.Pointer(&value))
	header.Cap, header.Len, header.Data = int(len), int(len), uintptr(unsafe.Pointer(data))
	return value
}
