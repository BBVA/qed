package hyper

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
)

func TestAdd(t *testing.T) {

	hasher, n := hashing.Sha256Hasher()
	cache := storage.NewSimpleCache(5000)
	store := storage.NewBadgerStorage("/tmp/badger_test")
	ht := NewTree("my test tree", hasher, n, 30, cache, store)

	for i := 0; i < 5; i++ {

		event := fmt.Sprintf("Hello World%d", i)
		key := hasher([]byte(event))
		value := make([]byte, 8)
		binary.LittleEndian.PutUint64(value, uint64(i))
		commitment, err := ht.Add(key, value)
		if err != nil {
			t.Log(err)
		}
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
	hasher, digestLength := hashing.Sha256Hasher()
	store := storage.NewBadgerStorage("/tmp/badger_bench") // NewBoltStorage("/tmp/bolt_test.db"
	cache := storage.NewSimpleCache(50000000)
	ht := NewTree("my bench tree", hasher, digestLength, 30, cache, store) 
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
		value := randomBytes(1)
		store.Add(key, value)
		ht.Add(key, value)
	}
	b.Logf("stats = %+v\n", ht.stats)
}
