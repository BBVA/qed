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

// Hahser expose the function hash and its properties
type Hasher struct {
	Do     func(...[]byte) []byte	// computes the hash function
	Size   int	// the size in bits of the hash function
	Maxval []byte // the maximum value the has function can return
}

// Returns a SHA256 HashFunc
func Hash256() *Hasher {
	var h Hasher
	h.Do = func(data ...[]byte) []byte {
		hasher := sha256.New()

		for i := 0; i < len(data); i++ {
			hasher.Write(data[i])
		}

		return hasher.Sum(nil)[:]
	}
	h.Size = 256
	h.Maxval = []byte{
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	return  &h
}

func Pow(x, y uint64) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}
