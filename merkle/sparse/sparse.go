// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package sparse

import (
	"verifiabledata/store"
	"verifiabledata/tree"
	"verifiabledata/util"
)

// Tree implements a Sparse Merkle Tree as stated in the paper
// ....
type Tree struct {
	s store.Store // ordered storage
}

// NewTree returns an instance of an Sparse Merkle Tree as stated by
// http://.....
func NewTree(s store.Store) *Tree {
	return &Tree{
		s,
	}
}

func (t *Tree) Add(digest util.Digest, index tree.Index) (*tree.Node, error) {

	return nil, nil
}
