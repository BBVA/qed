/*
    Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/

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
