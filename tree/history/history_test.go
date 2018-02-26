// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package history

import (
	"crypto/rand"
	"fmt"
	"testing"
	"verifiabledata/store/memory"
)

func TestAdd(t *testing.T) {
	var testCases = []struct {
		index      uint64
		commitment string
		event      string
	}{
		{0, "5cd26c62ee55c4a327fc7ec1eae97a232e7355f4340adfb0b3ca25b8d94135bd", "Hello World1"},
		{0, "81d3aa6da152370015e028ef97e9d303ffbf7ae121e362059e66bd217d5e09ce", "Hello World2"},
		{0, "0871c0d34eb2311a101cf1de957d15103c014885b1c306354766fbca2bc3d10e", "Hello World3"},
		{0, "8e4d915dcdbe9fd485336ecb7fa6780fc901179c6c5ded78781661120f3e3365", "Hello World4"},
		{0, "377f2fb38a02913effc8ec6de5bf51bfe1ebe2e473ea4fb5060f94b7c11b676e", "Hello World5"},
	}

	frozen := memory.NewStore()
	events := memory.NewStore()
	ht := NewTree(frozen, events)

	for _, e := range testCases {
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

// BenchmarkAdd-4   	  200000	      8166 ns/op
func BenchmarkAdd(b *testing.B) {
	frozen := memory.NewStore()
	events := memory.NewStore()
	ht := NewTree(frozen, events)
	data := make([]byte, 64)
	rand.Read(data)
	for i := 0; i < b.N; i++ {
		ht.Add(data)
	}
}
