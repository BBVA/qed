package sparse

import (
	"bytes"
	"math/big"
	"strconv"
)

type cache interface {
	// Exists checks if a value exists in the cache
	Exists(depth uint64, base *big.Int) bool
	// Get returns a value that exists from the cache
	Get(depth uint64, base *big.Int) []byte
	// Update maybe caches or clear the previous value
	Update(left, right []byte, depth uint64, base *big.Int,
		interiorHash []byte, defaultHashes [][]byte)
}

// CacheNothing is a naïve approach that caches nothing
type CacheNothing int

// Exists checks if a value exists in the cache
func (c CacheNothing) Exists(depth uint64, base *big.Int) bool {
	return false
}

// Get returns a value that exists from the cache
func (c CacheNothing) Get(depth uint64, base *big.Int) []byte {
	return nil
}

// Update maybe caches or clear the previous value
func (c CacheNothing) Update(left, right []byte, depth uint64, base *big.Int,
	interiorHash []byte, defaultHashes [][]byte) {
	// ignore
}

// CacheBranch caches every branch where both children have non-default values
type CacheBranch map[string][]byte

// Exists checks if a value exists in the cache
func (c CacheBranch) Exists(depth uint64, base *big.Int) bool {
	_, exists := c[strconv.Itoa(int(depth))+base.String()]
	return exists
}

// Get returns a value that exists from the cache
func (c CacheBranch) Get(depth uint64, base *big.Int) []byte {
	return c[strconv.Itoa(int(depth))+base.String()]
}

// Update maybe caches or clear the previous value
func (c CacheBranch) Update(left, right []byte, depth uint64, base *big.Int,
	interiorHash []byte, defaultHashes [][]byte) {
	if !bytes.Equal(defaultHashes[depth-1], left) && !bytes.Equal(defaultHashes[depth-1], right) {
		c[strconv.Itoa(int(depth))+base.String()] = interiorHash // update
	} else {
		delete(c, strconv.Itoa(int(depth))+base.String()) // clear previous
	}
}
