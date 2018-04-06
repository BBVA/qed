package history

// stats to measure the performance of the tree and
// the associated caches and strategies
type stats struct {
	unfreezing     int
	unfreezingHits int
	freezing       int
	leafHashes     int
	internalHashes int
}
