package common

import (
	"github.com/bbva/qed/db"
)

type Cache interface {
	Get(pos Position) (Digest, bool)
}

type ModifiableCache interface {
	Put(pos Position, value Digest)
	Fill(r db.KVPairReader) error
	Cache
}
type PassThroughCache struct {
	prefix byte
	store  db.Store
}

func NewPassThroughCache(prefix byte, store db.Store) *PassThroughCache {
	return &PassThroughCache{prefix, store}
}

func (c PassThroughCache) Get(pos Position) (Digest, bool) {
	pair, err := c.store.Get(c.prefix, pos.Bytes())
	if err != nil {
		return nil, false
	}
	return pair.Value, true
}

const keySize = 34

type SimpleCache struct {
	cached map[[keySize]byte]Digest
}

func NewSimpleCache(initialSize uint64) *SimpleCache {
	return &SimpleCache{make(map[[keySize]byte]Digest, initialSize)}
}

func (c SimpleCache) Get(pos Position) (Digest, bool) {
	var key [keySize]byte
	copy(key[:], pos.Bytes())
	digest, ok := c.cached[key]
	return digest, ok
}

func (c *SimpleCache) Put(pos Position, value Digest) {
	var key [keySize]byte
	copy(key[:], pos.Bytes())
	c.cached[key] = value
}

func (c *SimpleCache) Fill(r db.KVPairReader) (err error) {
	defer r.Close()
	for {
		entries := make([]*db.KVPair, 100)
		n, err := r.Read(entries)
		if err != nil || n == 0 {
			break
		}
		for _, entry := range entries {
			var key [keySize]byte
			copy(key[:], entry.Key)
			c.cached[key] = entry.Value
		}
	}
	return nil
}
