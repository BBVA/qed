/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

package hyper

import (
	"bytes"
	"sort"
)

// Cache interface defines the operations a cache mechanism must implement to
// be usable within the tree
type Cache interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, bool)
	Exists(key []byte) bool
	Size() uint64
}

// Store interface defines what the trees expect from a storage engine to implement
type Store interface {
	Add(key []byte, value []byte) error
	GetRange(start, end []byte) [][]byte
	Get(key []byte) ([]byte, error)
	Close() error
}

type DeletableStore interface {
	Delete(key []byte) error

	Store
}

// Split splits the slice.
func Split(ls [][]byte, s []byte) (left, right [][]byte) {
	// the smallest index i where d[i] >= s
	i := sort.Search(len(ls), func(i int) bool {
		return bytes.Compare(ls[i], s) >= 0
	})
	return ls[:i], ls[i:]
}
