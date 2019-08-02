/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package rocksdb

import (
	"sync"
	"sync/atomic"
)

// COWList implements a copy-on-write list. It is intended to be used by go
// callback registry for CGO, which is read-heavy with occasional writes.
// Reads do not block; Writes do not block reads (or vice versa), but only
// one write can occur at once;
type COWList struct {
	v  *atomic.Value
	mu *sync.Mutex
}

// NewCOWList creates a new COWList.
func NewCOWList() *COWList {
	var list []interface{}
	v := &atomic.Value{}
	v.Store(list)
	return &COWList{v: v, mu: new(sync.Mutex)}
}

// Append appends an item to the COWList and returns the index for that item.
func (c *COWList) Append(i interface{}) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	list := c.v.Load().([]interface{})
	newLen := len(list) + 1
	newList := make([]interface{}, newLen)
	copy(newList, list)
	newList[newLen-1] = i
	c.v.Store(newList)
	return newLen - 1
}

// Get gets the item at index.
func (c *COWList) Get(index int) interface{} {
	list := c.v.Load().([]interface{})
	return list[index]
}

type Registry struct {
	data map[string]interface{}
	sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		data: make(map[string]interface{}, 0),
	}
}

func (r *Registry) Register(key string, elem interface{}) {
	r.Lock()
	defer r.Unlock()
	r.data[key] = elem
}

func (r *Registry) Lookup(key string) interface{} {
	r.Lock()
	defer r.Unlock()
	return r.data[key]
}

func (r *Registry) Unregister(key string) {
	r.Lock()
	defer r.Unlock()
	delete(r.data, key)
}