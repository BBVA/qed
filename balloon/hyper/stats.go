package hyper

// stats to measure the performance of the tree and
// the associated caches and strategies
type stats struct {
	hits   int
	disk   int
	dh     int
	update int
	leaf   int
	lh     int
	ih     int
	lend   float64
}

