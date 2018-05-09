package storage

// INFO: 2^30 == 1073741824 nodes, 30 levels of cache
const SIZE30 = 1073741824

// INFO: 2^20 == 1048575 nodes, 20 levels of cache
const SIZE20 = 1048575

// INFO: 2^25 == 33554432 nodes, 25 levels of cache
const SIZE25 = 33554432

// Cache interface defines the operations a cache mechanism must implement to
// be usable within the tree
type Cache interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, bool)
	Exists(key []byte) bool
	Size() uint64
}
