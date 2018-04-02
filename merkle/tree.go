package merkle

// Tree defines the common API of any Merkle tree implementation.
type Tree interface {
	// Add inserts a new key-value pair into the tree and returns the
	// root hash as a commitment.
	Add(key []byte, value []byte) []byte
}
