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

package history

import (
	"github.com/BBVA/qed/balloon/hashing"
)

// Constant Zero is the 0x0 byte, and it is used as a prefix to know
// if a node has a zero digest.
var Zero = []byte{0x0}

// Constant One is the 0x1 byte, and it is used as a prefix to know
// if a node has a non-zero digest.
var One = []byte{0x1}

type leafHasher func([]byte, []byte) []byte
type interiorHasher func([]byte, []byte, []byte) []byte

func leafHasherF(hasher hashing.Hasher) leafHasher {
	return func(a, key []byte) []byte {
		return hasher(a, key)
	}
}

func interiorHasherF(hasher hashing.Hasher) interiorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(a, left, right)
	}
}
