package bolt

import (
	"crypto/rand"
	"fmt"
	"os"
)

func openBoltStorage() (*BoltStorage, func()) {
	store := NewBoltStorage("/tmp/bolt_store_test.db", "test")
	return store, func() {
		store.Close()
		deleteFile("/tmp/bolt_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
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
