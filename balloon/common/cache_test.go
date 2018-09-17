package common

import (
	"testing"

	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/util"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/testutils/rand"
	storage_utils "github.com/bbva/qed/testutils/storage"
)

func TestPassThroughCache(t *testing.T) {

	testCases := []struct {
		pos    Position
		value  hashing.Digest
		cached bool
	}{
		{&FakePosition{[]byte{0x0}, 0}, hashing.Digest{0x1}, true},
		{&FakePosition{[]byte{0x1}, 0}, hashing.Digest{0x2}, true},
		{&FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x3}, false},
	}

	store, closeF := storage_utils.OpenBPlusTreeStore()
	defer closeF()
	prefix := byte(0x0)
	cache := NewPassThroughCache(prefix, store)

	for i, c := range testCases {
		if c.cached {
			err := store.Mutate([]storage.Mutation{
				{prefix, c.pos.Bytes(), c.value},
			})
			require.NoError(t, err)
		}

		cachedValue, ok := cache.Get(c.pos)

		if c.cached {
			require.Truef(t, ok, "The key should exists in cache in test case %d", i)
			require.Equalf(t, c.value, cachedValue, "The cached value should be equal to stored value in test case %d", i)
		} else {
			require.Falsef(t, ok, "The key should not exist in cache in test case %d", i)
		}
	}

}

func TestSimpleCache(t *testing.T) {

	testCases := []struct {
		pos    Position
		value  hashing.Digest
		cached bool
	}{
		{&FakePosition{[]byte{0x0}, 0}, hashing.Digest{0x1}, true},
		{&FakePosition{[]byte{0x1}, 0}, hashing.Digest{0x2}, true},
		{&FakePosition{[]byte{0x2}, 0}, hashing.Digest{0x3}, false},
	}

	cache := NewSimpleCache(0)

	for i, c := range testCases {
		if c.cached {
			cache.Put(c.pos, c.value)
		}

		cachedValue, ok := cache.Get(c.pos)

		if c.cached {
			require.Truef(t, ok, "The key should exists in cache in test case %d", i)
			require.Equalf(t, c.value, cachedValue, "The cached value should be equal to stored value in test case %d", i)
		} else {
			require.Falsef(t, ok, "The key should not exist in cache in test case %d", i)
		}
	}
}

func TestFillSimpleCache(t *testing.T) {

	numElems := uint64(10000)
	cache := NewSimpleCache(0)
	reader := NewFakeKVPairReader(numElems)

	err := cache.Fill(reader)

	require.NoError(t, err)
	require.Truef(t, reader.Remaining == 0, "All elements should be cached. Remaining: %d", reader.Remaining)

	for i := uint64(0); i < numElems; i++ {
		pos := &FakePosition{util.Uint64AsBytes(i), 0}
		_, ok := cache.Get(pos)
		require.Truef(t, ok, "The element in position %v should be in cache", pos)
	}
}

type FakeKVPairReader struct {
	Remaining uint64
	index     uint64
}

func NewFakeKVPairReader(numElems uint64) *FakeKVPairReader {
	return &FakeKVPairReader{numElems, 0}
}

func (r *FakeKVPairReader) Read(buffer []*storage.KVPair) (n int, err error) {
	for n = 0; r.Remaining > 0 && n < len(buffer); n++ {
		pos := &FakePosition{util.Uint64AsBytes(r.index), 0}
		buffer[n] = &storage.KVPair{pos.Bytes(), rand.Bytes(8)}
		r.Remaining--
		r.index++
	}
	return n, nil
}
func (r *FakeKVPairReader) Close() {
	r.Remaining = 0
}
