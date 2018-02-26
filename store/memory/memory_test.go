// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package memory

import (
	"bytes"
	"testing"
	"verifiabledata/tree"
	"verifiabledata/util"
)

func TestAdd(t *testing.T) {
	s := NewStore()

	node := &tree.Node{&tree.Position{0, 0}, util.Hash([]byte("Hello World1"))}
	err := s.Add(node)
	if err != nil {
		t.Fatal("Error adding node to memory store")
	}

}

func TestGet(t *testing.T) {
	s := NewStore()

	node := &tree.Node{&tree.Position{0, 0}, util.Hash([]byte("Hello World1"))}
	err := s.Add(node)
	if err != nil {
		t.Fatal("Error adding node to memory store")
	}
	newNode, err := s.Get(node.Pos)
	if err != nil {
		t.Fatal("Error getting node from memory store")
	}
	if node.Pos.Index != newNode.Pos.Index || node.Pos.Layer != newNode.Pos.Layer || !bytes.Equal(node.Digest, newNode.Digest) {
		t.Fatal("Node is different than newnode: ", node, newNode)
	}
}
