package hyper


// A position identifies a unique node in the tree by its base, split and height
type Position struct {
	base   []byte // the left-most leaf in this node subtree
	split  []byte // the left-most leaf in the right branch of this node subtree
	height int    // height in the tree of this node
	n      int    // number of bits in the hash key
}

// returns a string representation of the position
func (p Position) String() string {
	return string(p.base[:byteslen]) + strconv.Itoa(p.height)
	// return fmt.Sprintf("%x-%d", p.base, p.height)
}

// returns a new position pointing to the left child
func (p Position) left() *Position {
	var np Position
	np.base = p.base
	np.height = p.height - 1
	np.n = p.n

	np.split = make([]byte, byteslen)
	copy(np.split, np.base)

	bitSet(np.split, p.n-p.height)

	return &np
}

// returns a new position pointing to the right child
func (p Position) right() *Position {
	var np Position
	np.base = p.split
	np.height = p.height - 1
	np.n = p.n

	np.split = make([]byte, byteslen)
	copy(np.split, np.base)

	bitSet(np.split, p.n-p.height)

	return &np
}

// creates the tree root position
func rootpos(n int) *Position {
	var p position
	p.base = make([]byte, byteslen)
	p.split = make([]byte, byteslen)
	p.height = n
	p.n = n

	bitSet(p.split, 0)

	return &p
}



func bitSet(bits []byte, i int)   { bits[i/8] |= 1 << uint(7-i%8) }
func bitUnset(bits []byte, i int) { bits[i/8] &= 0 << uint(7-i%8) }
