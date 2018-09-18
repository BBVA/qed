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

package util

import "encoding/binary"

func Uint64AsBytes(i uint64) []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, i)
	return b
}

func Uint16AsBytes(i uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return b
}

func BytesAsUint64(b []byte) uint64 {
	var out uint64
	for i, x := range b {
		out |= uint64(x) << uint64(i*8)
	}
	return out
}
