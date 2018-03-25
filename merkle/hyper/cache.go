package hyper

// Cache interface defines the operations a cache mechanism must implement to
// be usabel within the tree
type Cache interface {
	Insert(*position,[]bte) error
	Has(*position) bool
}

// area of the tree designated by its min and max height
type Area struct {
	minHeigth, maxHeigth int
}

// check if a position is whithing the caching area
func (a area) has(p *Position) bool {
	if p.height > min && p.height < max {
		return true
	}
	return false
}

// posssible overflow
func (a area) size() uint64 {
	return 2 ^ (a.max + 1) - 1 - 2 ^ (a.min + 1) - 1
}

// creates a new area structure, initialized with max and min boundaries
func newarea(min, max int) *area {
	return &area{
		min,
		max,
	}
}

// a cache contains the hashes of the pre computed nodes
type SimpleCache struct {
	n    int               // number of bits in the hash key
	node map[string][]byte // node map containing the cached hashes
	area *area             // min height of the cache
}


func (c* simplecache) Insert(p *position, h []byte) {
	if c.area.has(p) {
		c.node[p.String()] = nh
	}
}

func (c *simplecache) Has(p *position) bool {
	return c.area.has(p)
}

// creates a new cache structure, already initialized with
func NewSimpleCache(a *area, n int) *Cache {
	return &SimpleCache{
		n,
		make(map[string][]byte, a.size()),
		a,
	}
}


