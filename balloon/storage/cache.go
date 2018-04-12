// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package storage

// Cache interface defines the operations a cache mechanism must implement to
// be usable within the tree
type Cache interface {
	Put(key []byte, value []byte) error
	Get(key []byte) ([]byte, bool)
	Exists(key []byte) bool
}
