// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package sparse

import (
	"fmt"
	"testing"
	"verifiabledata/util"
)

func TestAdd(t *testing.T) {
	var testCases = []struct {
		v     uint64
		h     string
		event string
	}{
		{0, "5cd26c62ee55c4a327fc7ec1eae97a232e7355f4340adfb0b3ca25b8d94135bd", "Hello World1"},
		{1, "81d3aa6da152370015e028ef97e9d303ffbf7ae121e362059e66bd217d5e09ce", "Hello World2"},
		{2, "0871c0d34eb2311a101cf1de957d15103c014885b1c306354766fbca2bc3d10e", "Hello World3"},
		{3, "8e4d915dcdbe9fd485336ecb7fa6780fc901179c6c5ded78781661120f3e3365", "Hello World4"},
		{4, "377f2fb38a02913effc8ec6de5bf51bfe1ebe2e473ea4fb5060f94b7c11b676e", "Hello World5"},
	}

	leaves := NewInmemoryStore()
	cache := new(CacheBranch)

	st := NewTree("test tree", leaves, cache, util.Hash256())

	for _, e := range testCases {
		t.Log("Testing event: ", e.event)
		commitment, err := st.Add([]byte(e.h), e.v)
		if err != nil {
			t.Fatal("Error in Add call: ", err)
		}

		c := fmt.Sprintf("%x", commitment)
		if e.h != c {
			t.Fatal("Incorrect commitment: ", e.h, " != ", c)
		}
	}
}
