// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file
package history

import (
	"verifiabledata/balloon/hashing"
)

func fakeLeafHasherF(hasher hashing.Hasher) LeafHasher {
	return func(a, key []byte) []byte {
		return hasher(a, key)
	}
}

func fakeInteriorHasherF(hasher hashing.Hasher) InteriorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(a, left, right)
	}
}

func fakeLeafHasherCleanF(hasher hashing.Hasher) LeafHasher {
	return func(a, key []byte) []byte {
		return hasher(key)
	}
}

func fakeInteriorHasherCleanF(hasher hashing.Hasher) InteriorHasher {
	return func(a, left, right []byte) []byte {
		return hasher(left, right)
	}
}