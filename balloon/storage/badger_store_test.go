package storage

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func TestAdd(t *testing.T) {
	store, closeF := openStorage()
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
	store, closeF := openStorage()
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
	store, closeF := openStorage()
	defer closeF()

	key := []byte("Key")

	_, err := store.Get(key)
	if err == nil {
		t.Fatalf("Expected exception when trying to return a non-existent key")
	}

}

func TestGetRange(t *testing.T) {
	store, closeF := openStorage()
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

func openStorage() (*BadgerStorage, func()) {
	store := NewBadgerStorage("/tmp/badger_store_test.db")
	return store, func() {
		fmt.Println("Cleaning...")
		store.Close()
		deleteFile("/tmp/badger_store_test.db")
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
