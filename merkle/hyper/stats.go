package hyper

// stats to measure the performance of the tree and
// the associated caches and strategies
type stats struct {
	hits   uint64
	disk   uint64
	dh     uint64
	update uint64
	leaf   uint64
	lh     uint64
	ih     uint64
	lend   float64
}

