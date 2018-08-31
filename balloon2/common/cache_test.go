package common

import (
	"testing"

	"github.com/bbva/qed/db"
	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/testutils/storage"
)

func TestPassThroughCache(t *testing.T) {

	testCases := []struct {
		pos    Position
		value  Digest
		cached bool
	}{
		{&FakePosition{[]byte{0x0}, 0}, Digest{0x1}, true},
		{&FakePosition{[]byte{0x1}, 0}, Digest{0x2}, true},
		{&FakePosition{[]byte{0x2}, 0}, Digest{0x3}, false},
	}

	store, closeF := storage.NewBPlusTreeStore()
	defer closeF()
	prefix := byte(0x0)
	cache := NewPassThroughCache(prefix, store)

	for i, c := range testCases {
		if c.cached {
			err := store.Mutate(*db.NewMutation(prefix, c.pos.Bytes(), c.value))
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
		value  Digest
		cached bool
	}{
		{&FakePosition{[]byte{0x0}, 0}, Digest{0x1}, true},
		{&FakePosition{[]byte{0x1}, 0}, Digest{0x2}, true},
		{&FakePosition{[]byte{0x2}, 0}, Digest{0x3}, false},
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
