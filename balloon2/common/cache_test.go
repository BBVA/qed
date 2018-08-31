package common

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/bbva/qed/db"
	"github.com/bbva/qed/util"
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

type FakePosition struct {
	index  []byte
	height uint16
}

func (p FakePosition) Index() []byte {
	return p.index
}

func (p FakePosition) Height() uint16 {
	return p.height
}

func (p FakePosition) Bytes() []byte {
	b := make([]byte, 34) // Size of the index plus 2 bytes for the height
	copy(b, p.index)
	copy(b[len(p.index):], util.Uint16AsBytes(p.height))
	return b
}

func (p FakePosition) String() string {
	return fmt.Sprintf("Pos(%x, %d)", p.index, p.height)
}

func (p FakePosition) StringId() string {
	return fmt.Sprintf("%x|%d", p.index, p.height)
}

func (p FakePosition) IndexAsUint64() uint64 {
	return binary.BigEndian.Uint64(p.index)
}
