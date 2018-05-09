package cache

import (
	"bytes"
	"testing"
)

func TestPut(t *testing.T) {
	cache := NewSimpleCache(10)
	key := []byte("Key")
	value := []byte("Value")

	cache.Put(key, value)
	cached, ok := cache.Get(key)
	if !ok {
		t.Fatalf("Key not found: %d", key)
	}
	if bytes.Compare(cached, value) != 0 {
		t.Fatalf("Cached value [%d] does not match with original [%d]", cached, value)
	}
}

func TestExists(t *testing.T) {
	cache := NewSimpleCache(10)
	key := []byte("Key")
	value := []byte("Value")

	cache.Put(key, value)
	if !cache.Exists(key) {
		t.Fatalf("Cached key should exist: %d", key)
	}
}

func TestNotExists(t *testing.T) {
	cache := NewSimpleCache(10)
	key := []byte("Key")
	if cache.Exists(key) {
		t.Fatalf("Key should not exist in cache: %d", key)
	}
}

func TestGet(t *testing.T) {
	cache := NewSimpleCache(10)
	key := []byte("Key")
	value := []byte("Value")

	cache.Put(key, value)
	cached, ok := cache.Get(key)
	if cached == nil {
		t.Fatalf("Cached key should exist: %d", key)
	}
	if !ok {
		t.Fatalf("Cached key should exist: %d", key)
	}
	if bytes.Compare(cached, value) != 0 {
		t.Fatalf("Cached value [%d] does not match with original [%d]", cached, value)
	}
}

func TestGetNotCached(t *testing.T) {
	cache := NewSimpleCache(10)
	key := []byte("Key")

	cached, ok := cache.Get(key)
	if cached != nil {
		t.Fatalf("Cached key should not exist: %d", key)
	}
	if ok {
		t.Fatalf("Cached key should not exist: %d", key)
	}
}
