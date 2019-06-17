package cache

import (
	"container/list"

	"github.com/bbva/qed/storage"
)

const lruKeySize = 10

type entry struct {
	key   [lruKeySize]byte
	value []byte
}

// LruReadThroughCache implemets a "Last Recent Used" cache with a storage backend.
// Therefore, "writes" are done in-memory, but "reads" are done checking the storage.
// It uses an "evictList" for the LRU functionality.
type LruReadThroughCache struct {
	table     storage.Table
	store     storage.Store
	size      int
	items     map[[lruKeySize]byte]*list.Element
	evictList *list.List
}

// NewLruReadThroughCache returns a new cacheSize cache with an asociated storage.
func NewLruReadThroughCache(table storage.Table, store storage.Store, cacheSize uint16) *LruReadThroughCache {
	return &LruReadThroughCache{
		table:     table,
		store:     store,
		size:      int(cacheSize),
		items:     make(map[[lruKeySize]byte]*list.Element),
		evictList: list.New(),
	}
}

// Get function returns the value of a given key in cache. If it is not present, it looks
// on the storage. Finally it returns a boolean showing if the key is or is not present.
func (c LruReadThroughCache) Get(key []byte) ([]byte, bool) {
	var k [lruKeySize]byte
	copy(k[:], key)
	e, ok := c.items[k]
	if !ok {
		pair, err := c.store.Get(c.table, key)
		if err != nil {
			return nil, false
		}
		return pair.Value, true
	}
	c.evictList.MoveToFront(e)
	return e.Value.(*entry).value, ok
}

// Put function adds a new key/value element to the in-memory cache, or updates it if it
// exists. It also updates the LRU eviction list.
func (c *LruReadThroughCache) Put(key []byte, value []byte) {
	var k [lruKeySize]byte
	copy(k[:], key)
	// check for existing item
	if e, ok := c.items[k]; ok {
		// update value for specified key
		c.evictList.MoveToFront(e)
		e.Value.(*entry).value = value
		return
	}

	// Add new item
	e := &entry{k, value}
	entry := c.evictList.PushFront(e)
	c.items[k] = entry

	// Verify if eviction is needed
	if c.evictList.Len() > c.size {
		c.removeOldest()
	}
}

// Fill function inserts a bulk of key/value elements into the cache.
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

// Size function returns the number of items currently in the cache.
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
