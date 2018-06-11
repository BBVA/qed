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

// Package storage implements the Cache public interface.
//
// It also defines some constants with predefined cache sizes.
package storage

// SIZE20 is the amount of nodes needed for a 20-level cache
const SIZE20 = 1 << 20

// SIZE25 is the amount of nodes needed for a 25-level cache
const SIZE25 = 1 << 25

// SIZE30 is the amount of nodes needed for a 30-level cache
const SIZE30 = 1 << 30

// Cache interface defines the operations a cache mechanism must implement to
// be usable within the tree
type Cache interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, bool)
	Exists(key []byte) bool
	Size() uint64
}
