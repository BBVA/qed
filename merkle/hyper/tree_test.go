package hyper

import (
	"crypto/rand"
	"verifiabledata/util"
	// 	"fmt"
	"testing"
)

/*
func TestUpdate(t *testing.T) {

	values := []value{
		{[]byte("5cd26c62ee55c4a327fc7ec1eae97a232e7355f4340adfb0b3ca25b8d94135bd"), []byte{0x01}},
		{[]byte("81d3aa6da152370015e028ef97e9d303ffbf7ae121e362059e66bd217d5e09ce"), []byte{0x02}},
		{[]byte("0871c0d34eb2311a101cf1de957d15103c014885b1c306354766fbca2bc3d10e"), []byte{0x03}},
		{[]byte("8e4d915dcdbe9fd485336ecb7fa6780fc901179c6c5ded78781661120f3e3365"), []byte{0x04}},
		{[]byte("377f2fb38a02913effc8ec6de5bf51bfe1ebe2e473ea4fb5060f94b7c11b676e"), []byte{0x05}},
	}

	ht := newtree("my test tree")
	for _, v := range values {
		commitment := ht.toCache(&v, rootpos(ht.hasher.Size))
		fmt.Printf("%x\n", commitment)
	}

}
*/

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BenchmarkAdd(b *testing.B) {
	ht := NewTree("my bench tree", util.Hash256(), 30, NewSimpleCache(50000000), NewBPlusTreeStorage())
	b.N = 10000
	for i := 0; i < b.N; i++ {
		ht.Add(randomBytes(64), randomBytes(1))
	}
	b.Logf("stats = %+v\n", ht.stats)
}
