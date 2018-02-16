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
	commitment string
	event  string
}{
	{0, "5cd26c62ee55c4a327fc7ec1eae97a232e7355f4340adfb0b3ca25b8d94135bd", "Hello World1"},
	{0, "81d3aa6da152370015e028ef97e9d303ffbf7ae121e362059e66bd217d5e09ce", "Hello World2"},
	{0, "0871c0d34eb2311a101cf1de957d15103c014885b1c306354766fbca2bc3d10e", "Hello World3"},
	{0, "0871c0d34eb2311a101cf1de957d15103c014885b1c306354766fbca2bc3d10e", "Hello World4"},
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
		if e.commitment != commitment {
			t.Fatal("Incorrect commitment: ", e.commitment, " != ", commitment)
		}
	}
}
