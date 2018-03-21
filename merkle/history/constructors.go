// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file
package history

import (
	"verifiabledata/util"
)

// Default constructor that creates a in memory storage for frozen and events
// storages.
// As default it uses a SHA256 HashFunc
func NewInmemoryTree() *Tree {
	frozen := NewInmemoryStore()
	events := NewInmemoryStore()
	return NewTree(frozen, events, util.Hash256())
}
