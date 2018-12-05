package cache

import (
	"container/list"

	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/storage"
)

const lruKeySize = 10

type entry struct {
	key   [lruKeySize]byte
	value hashing.Digest
}

type LruReadThroughCache struct {
	prefix    byte
	store     storage.Store
	size      int
	items     map[[lruKeySize]byte]*list.Element
	evictList *list.List
}

func NewLruReadThroughCache(prefix byte, store storage.Store, cacheSize uint16) *LruReadThroughCache {
	return &LruReadThroughCache{
		prefix:    prefix,
		store:     store,
		size:      int(cacheSize),
		items:     make(map[[lruKeySize]byte]*list.Element),
		evictList: list.New(),
	}
}

func (c LruReadThroughCache) Get(pos navigator.Position) (hashing.Digest, bool) {
	var key [lruKeySize]byte
	copy(key[:], pos.Bytes())
	e, ok := c.items[key]
	if !ok {
		pair, err := c.store.Get(c.prefix, pos.Bytes())
		if err != nil {
			return nil, false
		}
		return pair.Value, true
	}
	c.evictList.MoveToFront(e)
	return e.Value.(*entry).value, ok
}

func (c *LruReadThroughCache) Put(pos navigator.Position, value hashing.Digest) {
	var key [lruKeySize]byte
	copy(key[:], pos.Bytes())
	// check for existing item
	if e, ok := c.items[key]; ok {
		// update value for specified key
		c.evictList.MoveToFront(e)
		e.Value.(*entry).value = value
		return
	}

	// Add new item
	e := &entry{key, value}
	entry := c.evictList.PushFront(e)
	c.items[key] = entry

	// Verify if eviction is needed
	if c.evictList.Len() > c.size {
		c.removeOldest()
	}
}

func (c *LruReadThroughCache) Fill(r storage.KVPairReader) (err error) {
	defer r.Close()
	for {
		entries := make([]*storage.KVPair, 100)
		n, err := r.Read(entries)
		if err != nil || n == 0 {
			break
		}
		for _, e := range entries {
			if e != nil {
				var key [lruKeySize]byte
				copy(key[:], e.Key)
				// check for existing item
				if ent, ok := c.items[key]; ok {
					// update value for specified key
					c.evictList.MoveToFront(ent)
					ent.Value.(*entry).value = e.Value
					continue
				}

				// Add new item
				e := &entry{key, e.Value}
				entry := c.evictList.PushFront(e)
				c.items[key] = entry

				// Verify if eviction is needed
				if c.evictList.Len() > c.size {
					c.removeOldest()
				}
			}
		}
	}
	return nil
}

func (c *LruReadThroughCache) Size() int {
	return c.evictList.Len()
}

func (c *LruReadThroughCache) removeOldest() {
	e := c.evictList.Back()
	if e != nil {
		c.evictList.Remove(e)
		kv := e.Value.(*entry)
		delete(c.items, kv.key)
	}
}
