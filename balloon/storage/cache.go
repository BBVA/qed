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

package storage

// INFO: 2^30 == 1073741824 nodes, 30 levels of cache
const SIZE30 = 1073741824

// INFO: 2^20 == 1048575 nodes, 20 levels of cache
const SIZE20 = 1048575

// INFO: 2^25 == 33554432 nodes, 25 levels of cache
const SIZE25 = 33554432

// Cache interface defines the operations a cache mechanism must implement to
// be usable within the tree
type Cache interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, bool)
	Exists(key []byte) bool
	Size() uint64
}
