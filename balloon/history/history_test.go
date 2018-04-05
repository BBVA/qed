// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package history

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
	"verifiabledata/util"
)

func TestAdd(t *testing.T) {
	var testCases = []struct {
		index      uint64
		commitment string
		event      string
	}{
		{0, "b4aa0a376986b4ab072ed536d41a4df65de5d46da15ff8756bc7657da01d2f52", "Hello World1"},
		{1, "81d3aa6da152370015e028ef97e9d303ffbf7ae121e362059e66bd217d5e09ce", "Hello World2"},
		{2, "0871c0d34eb2311a101cf1de957d15103c014885b1c306354766fbca2bc3d10e", "Hello World3"},
		{3, "8e4d915dcdbe9fd485336ecb7fa6780fc901179c6c5ded78781661120f3e3365", "Hello World4"},
		{4, "377f2fb38a02913effc8ec6de5bf51bfe1ebe2e473ea4fb5060f94b7c11b676e", "Hello World5"},
	}

	hasher, _ := hashing.Sha256Hasher()
	ht := NewTree(storage.NewBadgerStorage("/tmp/httest.db"), hasher)

	for _, e := range testCases {
		t.Log("Testing event: ", e.event)
		commitment, err := ht.Add([]byte(e.event), util.UInt64AsBytes(e.index))
		if err != nil {
			t.Fatal("Error in Add call: ", err)
		}

		c := hex.EncodeToString(commitment)
		fmt.Println(c)
		fmt.Println(hex.EncodeToString(hasher([]byte{0x0}, []byte(e.event))))
		if e.commitment != c {
			t.Fatal("Incorrect commitment: ", e.commitment, " != ", c)
		}
	}
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
	hasher, _ := hashing.Sha256Hasher()
	ht := NewTree(storage.NewBadgerStorage("/tmp/htbenchmark.db"), hasher)
	N := 100000
	for i := 0; i < N; i++ {
		key := randomBytes(64)
		ht.Add(key, util.UInt64AsBytes(uint64(i)))
	}
}
