package hyper

import (
	"crypto/rand"
	"fmt"
	"testing"
)

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

func randomBytes(n int) []byte {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	return bytes
}

func BenchmarkAdd(b *testing.B) {
	ht := newtree("my bench tree")
	p := rootpos(ht.hasher.Size)
	b.N = 1000000
	for i := 0; i < b.N; i++ {
		ht.toCache(&value{randomBytes(64), randomBytes(1)}, p)
	}
	b.Log("cache hits ", ht.cache.hits, " cache miss ", ht.cache.miss, " default hash ", ht.cache.dh, " cache depth ", ht.cache.depth, " cache max depth ", ht.cache.maxDepth, " cache size ", len(ht.cache.node))
}
