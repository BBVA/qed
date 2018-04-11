package balloon

import (
	"crypto/rand"
	"fmt"
	"os"
	"testing"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bolt"
	"verifiabledata/balloon/storage/cache"
)

func TestAdd(t *testing.T) {
	path := "/tmp/testAdd"
	frozen := badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", path))
	leaves := badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", path))
	cache := cache.NewSimpleCache(5000000)
	balloon := NewBalloon("/tmp/testAdd", 5000000, hashing.Sha256Hasher, frozen, leaves, cache)
	defer balloon.history.Close()
	defer balloon.hyper.Close()
	defer deleteFilesInDir("/tmp/testAdd")

	var testCases = []struct {
		index         uint
		historyDigest string
		indexDigest   string
		event         string
	}{
		{1, "fa712069a4f6ece78619c7ab233b42b94e40a7bab384311ee1e16b101a8478ec", "ff6ea7855c5d2dde67d7cdf10d18d116457b9e9454e0f595de7611d69b9d301c", "Hello World1"},
		{2, "db1d613425a77f0f129c55af46407f74a804ac1fb9ea6b27694dbc3628bc299b", "173a3265185c0f6de406165a3505f09f271b6d15658b57b591f4df88d35605a7", "Hello World2"},
		{3, "8cacd440f12d6d5749d51ca6538680ea65117afaabf92d0c6d8cd4953f124a81", "4d2111d387095e4f7e35b10bdcec87cf8184161f6a7b67672d0d797b50af1447", "Hello World3"},
		{4, "9e9bd9f70aac3bd93dfa92ce049b4916581166f96bae32d422af4cefb0f24220", "84955497b5b43129b8570ccbb0cbf3240746cca8ee40b00c061315ab4c0a809c", "Hello World4"},
		{5, "cd477562ea3e395f387e91f6fc40efd56406b2332272f094830fb535deb1fc1e", "1c65d8ac3e65db7520fee5a81b11de675ef5b94a7f8ce8d8ca534733750e1156", "Hello World5"},
	}

	for _, e := range testCases {
		t.Log("Testing event: ", e.event)
		comm := balloon.Add([]byte(e.event))
		commitment := <- comm

		if e.index != commitment.Version {
			t.Fatal("Incorrect index: ", e.index, " != ", commitment.Version)
		}
		hd := fmt.Sprintf("%x", commitment.HistoryDigest)
		hyd := fmt.Sprintf("%x", commitment.IndexDigest)
		if e.historyDigest != hd {
			t.Fatal("Incorrect history commitment: ", e.historyDigest, " != ", hd)
		}

		if e.indexDigest != hyd {
			t.Fatal("Incorrect hyper commitment: ", e.indexDigest, " != ", hyd)
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

func deleteFilesInDir(path string) {
	os.RemoveAll(fmt.Sprintf("%s/leaves.db", path))
	os.RemoveAll(fmt.Sprintf("%s/frozen.db", path))
}

func BenchmarkAddBolt(b *testing.B) {
	path := "/tmp/benchAdd"
	frozen := bolt.NewBoltStorage(fmt.Sprintf("%s/frozen.db", path), "frozen")
	leaves := bolt.NewBoltStorage(fmt.Sprintf("%s/leaves.db", path), "leaves")
	cache := cache.NewSimpleCache(5000000)
	balloon := NewBalloon("/tmp/benchAdd", 5000000, hashing.Sha256Hasher, frozen, leaves, cache)
	defer deleteFilesInDir("/tmp/benchAdd")
	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := randomBytes(128)
		 r := balloon.Add(event)
		 <- r
	}

}

func BenchmarkAddBadger(b *testing.B) {
	path := "/tmp/benchAdd"
	frozen := badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", path))
	leaves := badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", path))
	cache := cache.NewSimpleCache(5000000)
	balloon := NewBalloon("/tmp/benchAdd", 5000000, hashing.Sha256Hasher, frozen, leaves, cache)
	defer deleteFilesInDir("/tmp/benchAdd")
	b.ResetTimer()
	b.N = 10000
	for i := 0; i < b.N; i++ {
		event := randomBytes(128)
		r := balloon.Add(event)
		<- r
	}

}
