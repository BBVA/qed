package hyper

import (
	"bytes"
	"testing"
)

func TestRoot(t *testing.T) {

	expectedBase := []byte{0x0}
	expectedSplit := []byte{0x80}

	root := rootPosition(8)

	if !bytes.Equal(root.base, expectedBase) {
		t.Fatalf("Wrong base: expected %b, actual %b", expectedBase, root.base)
	}

	if !bytes.Equal(root.split, expectedSplit) {
		t.Fatalf("Wrong split: expected %b, actual %b", expectedSplit, root.split)
	}

	if root.height != 8 {
		t.Fatalf("Wrong height: expected %v, actual %v", 8, root.height)
	}

	if root.n != 8 {
		t.Fatalf("Wrong n: expected %v, actual %v", 8, root.n)
	}

}
func TestRight(t *testing.T) {

	expectedBase := []byte{0x80}
	expectedSplit := []byte{0xC0}

	root := rootPosition(8)

	right := root.right()

	if !bytes.Equal(right.base, expectedBase) {
		t.Fatalf("Wrong base: expected %b, actual %b", expectedBase, right.base)
	}

	if !bytes.Equal(right.split, expectedSplit) {
		t.Fatalf("Wrong split: expected %b, actual %b", expectedSplit, right.split)
	}

	if right.height != root.height-1 {
		t.Fatalf("Wrong height: expected %v, actual %v", root.height-1, right.height)
	}

	if right.n != root.n {
		t.Fatalf("Wrong n: expected %v, actual %v", root.n, right.n)
	}

}

func TestLeft(t *testing.T) {

	expectedBase := []byte{0x00}
	expectedSplit := []byte{0x40}

	root := rootPosition(8)

	left := root.left()

	if !bytes.Equal(left.base, expectedBase) {
		t.Fatalf("Wrong base: expected %b, actual %b", expectedBase, left.base)
	}

	if !bytes.Equal(left.split, expectedSplit) {
		t.Fatalf("Wrong split: expected %b, actual %b", expectedSplit, left.split)
	}

	if left.height != root.height-1 {
		t.Fatalf("Wrong height: expected %v, actual %v", root.height-1, left.height)
	}

	if left.n != root.n {
		t.Fatalf("Wrong n: expected %v, actual %v", root.n, left.n)
	}

}
