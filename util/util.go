// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

/*
	Package util implements cross domain functions used all across the code
*/
package util

import (
	"crypto/sha256"
)

// The return type of the hash function, in our case a [32]byte
type Digest [sha256.Size]byte

// Returns a Digest of the []byte passed as a parameter
func Hash(buff []byte) Digest {
	return sha256.Sum256(buff)
}
