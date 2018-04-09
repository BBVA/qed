package hyper

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"os"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/cache"
)

func TestAdd(t *testing.T) {
	store, closeF := openStorage()
	defer closeF()

	cache := cache.NewSimpleCache(5000)

	ht := NewTree("my test tree", cache, store, hashing.Sha256Hasher)

	for i := 0; i < 5; i++ {
		event := fmt.Sprintf("Hello World%d", i)
		key := hashing.Sha256Hasher([]byte(event))
		value := make([]byte, 8)
		binary.LittleEndian.PutUint64(value, uint64(i))
		<-ht.Add(key, value)
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
	store, closeF := openStorage()
	defer closeF()
	cache := cache.NewSimpleCache(5000000)
	ht := NewTree("my test tree", cache, store, hashing.Sha256Hasher)
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
		value := randomBytes(1)
		store.Add(key, value)
		<-ht.Add(key, value)
	}
	b.Logf("stats = %+v\n", ht.stats)
}

func openStorage() (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage("/tmp/hyper_tree_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/tmp/hyper_tree_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
