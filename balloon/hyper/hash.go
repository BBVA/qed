// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package hyper

import (
	"bytes"
	"fmt"
	"runtime"
	"verifiabledata/balloon/hashing"
)

// Constant Empty is a constant for empty leaves
var Empty = []byte{0x00}

// Constant Set is a constant for non-empty leaves
var Set = []byte{0x01}

type LeafHasher func([]byte, []byte, []byte) []byte
type InteriorHasher func([]byte, []byte, []byte, []byte) []byte

func where(calldepth int) string {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		file = "???"
		line = 0
	}
	return fmt.Sprintf("%s:%d", file, line)
}

func LeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(id, a, base []byte) []byte {
		if bytes.Equal(a, Empty) {
			return hasher(id)
		}

		return hasher(id, base)
	}
}

func InteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(left, right, base, height []byte) []byte {
		if bytes.Equal(left, right) {
			return hasher(left, right)
		}

		return hasher(left, right, base, height)
	}
}
