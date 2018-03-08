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
	"math/big"
)

// Returns the hash of daa as an []byte, where data can
// be one or multiple []byte elements
type HashFunc func(data ...[]byte) []byte

// returns the hash results in big number format
func (h HashFunc) BigInt(data ...[]byte) *big.Int {
	return big.NewInt(0).SetBytes(h(data...))
}

// Return the size in bits of the hash function
// TODO Dangerous: this cannot be call often!!
func (h HashFunc) Size() int {
	return len(h([]byte("size"))) * 8
}

// Returns a SHA256 HashFunc 
func Hash256() HashFunc {
	return func(data ...[]byte) []byte {
		hasher := sha256.New()

		for i := 0; i < len(data); i++ {
			hasher.Write(data[i])
		}

		return hasher.Sum(nil)[:]
	}
}

func Pow(x, y uint64) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}
