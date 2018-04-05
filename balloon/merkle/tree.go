package merkle

// Tree defines the common API of any Merkle tree implementation.
type MerkleTree interface {
	// Add inserts a new key-value pair into the tree and returns the
	// root hash as a commitment.
	Add([]byte, []byte) ([]byte, error)
}

type TreeChannel struct {
	Send    <-chan *KVPair
	Receive chan<- []byte
	Signal  <-chan string
}

type KVPair struct {
	Digest []byte
	Index  []byte
}

func NewTreeChannel() *TreeChannel {
	return &TreeChannel{
		make(chan *KVPair),
		make(chan []byte),
		make(chan string),
	}
}
