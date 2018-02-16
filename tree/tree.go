// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

/*
	Package tree implements all the trees used in the system.
*/
package tree

import (
	"fmt"
	"verifiabledata/util"
)

// Data type holds the information of a leaf node
type Data []byte

// Commitment holds the Digest as proof of the event insertion, and the verstion of
// the tree after the insertion, which is equivalent to the Position.Index
type Commitment struct {
	Version uint64
	Digest  util.Digest
}

// Position holds the index and the layer of a node in a tree
type Position struct {
	Index uint64
	Layer uint64
}

func (p *Position) String() string {
	return fmt.Sprintf("(i %d, l %d)",p.Index, p.Layer)
}

// Returns a copy of position with l as Layer
func (p *Position) SetLayer(l uint64) *Position {
	return &Position{
		p.Index,
		l,
	}
}

// Returns a copy of position with i as Index
func (p *Position) SetIndex(i uint64) *Position {
	return &Position{
		i,
		p.Layer,
	}
}

// A node holds its digest and its position
type Node struct {
	Pos    *Position
	Digest util.Digest
}

func (n *Node) String() string {
	return fmt.Sprintf("(P %s,  D %x)", n.Pos, n.Digest)
}
