// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package history

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bplus"
)

func TestAdd(t *testing.T) {
	var testCases = []struct {
		index      uint64
		commitment []byte
		event      []byte
	}{
		{0, []byte{0x4a}, []byte{0x4a}},
		{1, []byte{0x00}, []byte{0x4b}},
		{2, []byte{0x48}, []byte{0x48}},
		{3, []byte{0x01}, []byte{0x49}},
	}

	store, closeF := openBPlusStorage()
	defer closeF()

	ht := NewTree(store, fakeLeafHasherF(hashing.XorHasher), fakeInteriorHasherF(hashing.XorHasher))

	for _, e := range testCases {
		t.Logf("Testing event: %b", e.event)
		commitment := <-ht.Add(e.event, uInt64AsBytes(uint(e.index)))

		if !bytes.Equal(e.commitment, commitment) {
			t.Fatalf("Incorrect commitment: expected %b, actual %b", e.commitment, commitment)
		}
	}
}

func TestProve(t *testing.T) {
	store, closeF := openBPlusStorage()
	defer closeF()

	var testCases = []struct {
		index      uint64
		commitment []byte
		event      []byte
	}{
		{0, []byte{0x4a}, []byte{0x4a}}, // 74
		{1, []byte{0x00}, []byte{0x4b}}, // 75
		{2, []byte{0x48}, []byte{0x48}}, // 72
		{3, []byte{0x01}, []byte{0x49}}, // 73
		{4, []byte{0x01}, []byte{0x50}}, // 80
		{5, []byte{0x01}, []byte{0x51}}, // 81
		{6, []byte{0x01}, []byte{0x52}}, // 82
	}

	ht := NewTree(store, fakeLeafHasherCleanF(hashing.XorHasher), fakeInteriorHasherCleanF(hashing.XorHasher))

	for _, e := range testCases {
		<-ht.Add(e.event, uInt64AsBytes(uint(e.index)))
	}

	expectedPath := [][]byte{
		[]byte{0x00},
		[]byte{0x52},
		[]byte{0x50},
	}
	proof := <-ht.Prove([]byte{0x5}, 6)

	fmt.Println(proof)
	if !comparePaths(expectedPath, proof.Nodes) {
		t.Fatalf("Invalid path: expected %v, actual %v", expectedPath, proof.Nodes)
	}
}

func comparePaths(expected [][]byte, actual []Node) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i, e := range expected {
		if !bytes.Equal(e, actual[i].Digest) {
			return false
		}
	}
	return true
}

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	ht := NewTree(store, LeafHasherF(hashing.Sha256Hasher), InteriorHasherF(hashing.Sha256Hasher))
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
		<-ht.Add(key, uInt64AsBytes(uint(i)))
	}
	b.Logf("stats = %+v\n", ht.stats)
}

func openBPlusStorage() (*bplus.BPlusTreeStorage, func()) {
	store := bplus.NewBPlusTreeStorage()
	return store, func() {
		store.Close()
	}
}

func openBadgerStorage() (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage("/tmp/history_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/tmp/history_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}

// Utility to generate graphviz code to visualize
// a tree
func graphTree(t *Tree, key []byte, version uint) {
	fmt.Println("digraph BST {")
	fmt.Println("	node [style=filled];")

	// start at root, and traverse the tree
	// using the frozen leaves
	// tree := new(repr)
	graphBody(t, version)
	//	fmt.Println(tree)

	fmt.Println("}")

}

func atob(a, b *node, color string) {
	fmt.Printf("	\"%s\" -> \"%s\" [%s];\n", a, b, color)
}

func strnode(digest []byte, index, layer uint) string {
	return fmt.Sprintf("[%x] (%d,%d)", digest, index, layer)
}

func strfroz(index, layer uint) string {
	return fmt.Sprintf("[??] (%d,%d)", index, layer)
}

type node struct {
	index, layer uint
	digest       string
}

func (n node) String() string {
	return fmt.Sprintf("[%s] (%d,%d)", n.digest, n.layer, n.index)
}

type stack []*node

func (s *stack) push(n *node) {
	*s = append(*s, n)
}
func (s *stack) pop() *node {
	first, rest := (*s)[0], (*s)[1:]
	*s = rest
	return first
}

func graphBody(t *Tree, version uint) error {
	var current *node
	var s *stack

	index := uint(0)
	layer := t.getDepth(version)
	digest, _ := t.frozen.Get(frozenKey(uint(index), uint(layer)))
	current = &node{index, layer, fmt.Sprintf("0x%x", digest)}
	s = &stack{}

	for {
		if current != nil {
			s.push(current)
			digest, _ = t.frozen.Get(frozenKey(index, layer-1))
			left := &node{current.index, current.layer - 1, fmt.Sprintf("0x%x", digest)}
			atob(current, left, "color=blue")
			if current.layer-1 == 0 {
				current = nil
			} else {
				current = left
			}
		} else {
			if len(*s) > 0 {
				current = s.pop()
				right := &node{current.index + pow(2, current.layer-1), current.layer - 1, "??"}
				atob(current, right, "color=green")
				if current.layer-1 == 0 {
					current = nil
				} else {
					current = right
				}
			} else {
				break
			}

		}

	}
	return nil
}
