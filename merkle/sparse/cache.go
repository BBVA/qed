package sparse

import (
	"fmt"
	"bytes"
	"strconv"
)

type cache interface {
	// Exists checks if a value exists in the cache
	Exists(depth int, base []byte) bool
	// Get returns a value that exists from the cache
	Get(depth int, base []byte) []byte
	// Update maybe caches or clear the previous value
	Update(left, right, base, interiorHash []byte, defaultHashes [][]byte, depth int)
}

// CacheNothing is a naïve approach that caches nothing
type CacheNothing int

// Exists checks if a value exists in the cache
func (c CacheNothing) Exists(depth uint64, base []byte) bool {
	return false
}

// Get returns a value that exists from the cache
func (c CacheNothing) Get(depth uint64, base []byte) []byte {
	return nil
}

// Update maybe caches or clear the previous value
func (c CacheNothing) Update(left, right []byte, depth uint64, base []byte,
	interiorHash []byte, defaultHashes [][]byte) {
	// ignore
}

// CacheBranch caches every branch where both children have non-default values
type CacheBranch map[string][]byte

// Exists checks if a value exists in the cache
func (c CacheBranch) Exists(depth int, base []byte) bool {
	b := fmt.Sprintf("%x",base)
	_, exists := c[strconv.Itoa(int(depth))+b]
	return exists
}

// Get returns a value that exists from the cache
func (c CacheBranch) Get(depth int, base []byte) []byte {
	b := fmt.Sprintf("%x",base)
	return c[strconv.Itoa(int(depth))+b]
}

// Update maybe caches or clear the previous value
func (c CacheBranch) Update(left, right, base, interiorHash []byte, defaultHashes [][]byte, depth int) {
	b := fmt.Sprintf("%x",base)
	if !bytes.Equal(defaultHashes[depth-1], left) && !bytes.Equal(defaultHashes[depth-1], right) {
		c[strconv.Itoa(int(depth))+b] = interiorHash // update
	} else {
		delete(c, strconv.Itoa(int(depth))+b) // clear previous
	}
}
