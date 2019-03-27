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

type LruReadThroughCache struct {
	table     storage.Table
	store     storage.Store
	size      int
	items     map[[lruKeySize]byte]*list.Element
	evictList *list.List
}

func NewLruReadThroughCache(table storage.Table, store storage.Store, cacheSize uint16) *LruReadThroughCache {
	return &LruReadThroughCache{
		table:     table,
		store:     store,
		size:      int(cacheSize),
		items:     make(map[[lruKeySize]byte]*list.Element),
		evictList: list.New(),
	}
}

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
