package history

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
)

func TestAdd(t *testing.T) {
	var testCases = []struct {
		index      uint64
		commitment string
		event      string
	}{
		{0, "b4aa0a376986b4ab072ed536d41a4df65de5d46da15ff8756bc7657da01d2f52", "Hello World1"},
		{1, "8c93bb614746a51c1200f01a0ba5217686bf576b8bb0b095523ea38c740c567e", "Hello World2"},
		{2, "1306d590e35d965aa42ca6e3b05b67cd009d7b9021f777c480e55eb626072dc4", "Hello World3"},
		{3, "d3e8bc7215dda0d39689a4bfc16974dd63e5420b35abb4860073dbbcb7e197ae", "Hello World4"},
		{4, "8b8e3177b98d00f6a9e6d621ac660331318524f5a0cee2a62472e8e8bf682fd8", "Hello World5"},
	}

	store, closeF := openStorage()
	defer closeF()

	ht := NewTree(store, hashing.Sha256Hasher)

	for _, e := range testCases {
		t.Log("Testing event: ", e.event)
		commitment := <-ht.Add([]byte(e.event), uInt64AsBytes(e.index))

		c := hex.EncodeToString(commitment)

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
	store, closeF := openStorage()
	defer closeF()
	ht := NewTree(store, hashing.Sha256Hasher)
	b.N = 100000
	for i := 0; i < b.N; i++ {
		key := randomBytes(64)
		<-ht.Add(key, uInt64AsBytes(uint64(i)))
	}
	b.Logf("stats = %+v\n", ht.stats)
}

func openStorage() (*badger.BadgerStorage, func()) {
	store := badger.NewBadgerStorage("/tmp/history_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/tmp/history_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
