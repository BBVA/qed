// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package history

import (
	"fmt"
	"testing"
	"verifiabledata/store/memory"
)

var eventTable = []struct {
	index  uint64
	digest string
	event  string
}{
	{0, "5cd26c62ee55c4a327fc7ec1eae97a232e7355f4340adfb0b3ca25b8d94135bd", "Hello World1"},
	{1, "975ab94d40843da8109a5e4c7d9577188fad3278dc4bc7ee576690171b8688f4", "Hello World2"},
	{2, "c15432da2b21261edf1c348b7cc290677aa58784de268fee4ab97e63689601ea", "Hello World3"},
	{3, "pepe", "Hello World4"},
}

func TestAdd(t *testing.T) {
	frozen := memory.NewMemoryStore()
	events := memory.NewMemoryStore()
	ht := NewHistoryTree(frozen, events)

	for _, e := range eventTable {
		t.Log("Testing event: ", e.event)
		node, err := ht.Add([]byte(e.event))
		if err != nil {
			t.Fatal("Error in Add call: ", err)
		}

		if e.index != node.Pos.Index {
			t.Fatal("Incorrect index: ", e.index, " != ", node.Pos.Index)
		}
		commitment := fmt.Sprintf("%x", node.Digest)
		if e.digest != commitment {
			t.Fatal("Incorrect commitment: ", e.digest, " != ", commitment)
		}
	}
}
