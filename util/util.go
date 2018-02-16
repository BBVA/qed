// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

/*
	Package util implements cross domain functions used all across the code
*/
package util

import (
	"crypto/sha256"
	"math"
)

// The return type of the hash function, in our case a [32]byte
type Digest []byte

// Returns a Digest of the []byte passed as a parameter
func Hash(buff []byte) Digest {
	d := sha256.Sum256(buff)
	return Digest(d[:])
}

func Pow(x, y uint64) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}
