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
		{4, []byte{0x4e}, []byte{0x4e}},
		{5, []byte{0x01}, []byte{0x4f}},
		{6, []byte{0x4c}, []byte{0x4c}},
		{7, []byte{0x01}, []byte{0x4d}},
		{8, []byte{0x43}, []byte{0x42}},
		{9, []byte{0x00}, []byte{0x43}},
	}

	store, closeF := openBPlusStorage()
	defer closeF()

	ht := NewFakeTree(store, hashing.XorHasher)

	for i, e := range testCases {
		commitment := <-ht.Add(e.event, uInt64AsBytes(e.index))

		if !bytes.Equal(e.commitment, commitment) {
			t.Fatalf("Incorrect commitment for test %d: expected %x, actual %x", i, e.commitment, commitment)
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

	ht := NewFakeTree(store, hashing.XorHasher)

	for _, e := range testCases {
		<-ht.Add(e.event, uInt64AsBytes(e.index))
	}

	expectedPath := [][]byte{
		[]byte{0x01},
		[]byte{0x00},
	}
	proof := <-ht.Prove([]byte{0x5}, 6, 6)
	graphTree(ht, []byte{0x5}, 6)
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
	ht := NewTree(store, hashing.Sha256Hasher)
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
		<-ht.Add(key, uInt64AsBytes(uint64(i)))
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
	store := badger.NewBadgerStorage("/var/tmp/history_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/var/tmp/history_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}

// Utility to generate graphviz code to visualize
// the frozen tree
func graphTree(t *Tree, key []byte, version uint64) {
	fmt.Println("digraph BST {")
	fmt.Println("	node [style=filled];")

	// start at root, and traverse the tree
	// using the frozen leaves
	//graphBody(t, version)
	graphTreePreoder(t, rootnode(t, version))

	//	fmt.Println(tree)

	fmt.Println("}")

}

func atob(a, b *node, color string) {
	fmt.Printf("	\"%s\" -> \"%s\" [%s];\n", a, b, color)
}

type node struct {
	index, layer uint64
	digest       string
}

func (n node) String() string {
	return fmt.Sprintf("[%s] (%d,%d)", n.digest, n.index, n.layer)
}

func rootnode(t *Tree, version uint64) *node {
	var index uint64 = 0
	layer := t.getDepth(version)
	digest, _ := t.frozen.Get(frozenKey(index, layer))
	return &node{index, layer, fmt.Sprintf("0x%x", digest)}
}

func graphTreePreoder(t *Tree, parent *node) {
	if parent.layer == 0 {
		return
	}

	digest, _ := t.frozen.Get(frozenKey(parent.index, parent.layer-1))
	left := &node{parent.index, parent.layer - 1, fmt.Sprintf("0x%x", digest)}
	atob(parent, left, "color=blue")
	graphTreePreoder(t, left)

	right := &node{parent.index + pow(2, parent.layer-1), parent.layer - 1, "??"}
	atob(parent, right, "color=green")
	graphTreePreoder(t, right)
}
