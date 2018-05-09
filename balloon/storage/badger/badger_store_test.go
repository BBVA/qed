// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package badger

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"testing"
)

func TestAdd(t *testing.T) {
	store, closeF := openBadgerStorage()
	defer closeF()

	key := []byte("Key")
	value := []byte("Value")

	err := store.Add(key, value)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.Get(key)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetExistentKey(t *testing.T) {
	store, closeF := openBadgerStorage()
	defer closeF()

	key := []byte("Key")
	value := []byte("Value")

	store.Add(key, value)

	stored, err := store.Get(key)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(stored, value) != 0 {
		t.Fatalf("The stored key does not match the original: expected %d, actual %d", value, stored)
	}

}

func TestNonExistentKey(t *testing.T) {
	store, closeF := openBadgerStorage()
	defer closeF()

	key := []byte("Key")

	value, _ := store.Get(key)
	if len(value) > 0 {
		t.Fatalf("Expected empty value")
	}

}

func TestGetRange(t *testing.T) {
	store, closeF := openBadgerStorage()
	defer closeF()

	var testValues = []struct {
		size  int
		start byte
		end   byte
	}{
		{40, 10, 50},
		{0, 1, 9},
		{11, 1, 20},
		{10, 40, 60},
		{0, 60, 100},
		{0, 20, 10},
	}

	for i := 10; i < 50; i++ {
		store.Add([]byte{byte(i)}, []byte("Value"))
	}

	for _, test := range testValues {
		slice := store.GetRange([]byte{test.start}, []byte{test.end})
		if len(slice) != test.size {
			t.Errorf("Slice length invalid: expected %d, actual %d", test.size, len(slice))
		}
	}

}

func TestDelete(t *testing.T) {
	store, closeF := openBadgerStorage()
	defer closeF()

	store.Add([]byte{0x0}, []byte{0x1})
	store.Delete([]byte{0x0})

	value, _ := store.Get([]byte{0x0})
	if value != nil {
		t.Errorf("Value should be nil, got '%v' instead", value)
	}
}

func BenchmarkAdd(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	b.N = 10000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.Add(randomBytes(128), []byte("Value"))
	}
}

func BenchmarkGet(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	N := 10000
	b.N = N
	var key []byte

	// populate storage
	for i := 0; i < N; i++ {
		if i == 10 {
			key = randomBytes(128)
			store.Add(key, []byte("Value"))
		} else {
			store.Add(randomBytes(128), []byte("Value"))
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.Get(key)
		if err != nil {
			b.Fatalf("Unexpected error: %s", err)
		}
	}

}

func BenchmarkGetRangeInLargeTree(b *testing.B) {
	store, closeF := openBadgerStorage()
	defer closeF()
	N := 1000000

	// populate storage
	for i := 0; i < N; i++ {
		store.Add([]byte{byte(i)}, []byte("Value"))
	}

	b.ResetTimer()

	b.Run("Small range", func(b *testing.B) {
		b.N = 10000
		for i := 0; i < b.N; i++ {
			slice := store.GetRange([]byte{10}, []byte{10})
			if len(slice) != 1 {
				b.Fatalf("Unexpected leaves slice size: %d", len(slice))
			}
		}
	})

	b.Run("Large range", func(b *testing.B) {
		b.N = 10000
		for i := 0; i < b.N; i++ {
			slice := store.GetRange([]byte{10}, []byte{35})
			if len(slice) != 26 {
				b.Fatalf("Unexpected leaves slice size: %d", len(slice))
			}
		}
	})

}

func openBadgerStorage() (*BadgerStorage, func()) {
	store := NewBadgerStorage("/var/tmp/badger_store_test.db")
	return store, func() {
		store.Close()
		deleteFile("/var/tmp/badger_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
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
