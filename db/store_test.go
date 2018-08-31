package db

import (
	"testing"

	"github.com/bbva/qed/util"
	"github.com/stretchr/testify/require"
)

func TestGetIndex(t *testing.T) {

	kvrange := NewKVRange()
	for i := uint64(0); i < 10; i++ {
		kvpair := NewKVPair(util.Uint64AsBytes(i), util.Uint64AsBytes(i))
		kvrange = kvrange.InsertSorted(kvpair)
	}

	tests := []struct {
		key           []byte
		expectedIndex int
	}{
		{util.Uint64AsBytes(uint64(3)), 3},
		{util.Uint64AsBytes(uint64(0)), 0},
		{util.Uint64AsBytes(uint64(9)), 9},
	}

	for i, test := range tests {
		require.Equalf(t, test.expectedIndex, getIndex(kvrange, test.key), "Error searching in test: %d, value: %x", i, test.key)
	}

}

func TestInsertSorted(t *testing.T) {

	kvrange := KVRange{
		NewKVPair([]byte{0x01}, []byte{0x01}),
		NewKVPair([]byte{0x08}, []byte{0x08}),
	}

	tests := []struct {
		item          KVPair
		expectedRange KVRange
	}{

		{NewKVPair([]byte{0x00}, []byte{0x00}),
			KVRange{
				NewKVPair([]byte{0x00}, []byte{0x00}),
				NewKVPair([]byte{0x01}, []byte{0x01}),
				NewKVPair([]byte{0x08}, []byte{0x08}),
			},
		},

		{NewKVPair([]byte{0x04}, []byte{0x04}),
			KVRange{
				NewKVPair([]byte{0x00}, []byte{0x00}),
				NewKVPair([]byte{0x01}, []byte{0x01}),
				NewKVPair([]byte{0x04}, []byte{0x04}),
				NewKVPair([]byte{0x08}, []byte{0x08}),
			},
		},

		{NewKVPair([]byte{0x09}, []byte{0x09}),
			KVRange{
				NewKVPair([]byte{0x00}, []byte{0x00}),
				NewKVPair([]byte{0x01}, []byte{0x01}),
				NewKVPair([]byte{0x04}, []byte{0x04}),
				NewKVPair([]byte{0x08}, []byte{0x08}),
				NewKVPair([]byte{0x09}, []byte{0x09}),
			},
		},
	}

	for i, test := range tests {
		kvrange = kvrange.InsertSorted(test.item)
		require.Equalf(t, test.expectedRange, kvrange, "Error sorting in test: %d, value: %x", i, test.item)
	}

}

func TestSplit(t *testing.T) {

	kvrange := KVRange{
		NewKVPair([]byte{0x00}, []byte{0x00}),
		NewKVPair([]byte{0x01}, []byte{0x01}),
		NewKVPair([]byte{0x02}, []byte{0x02}),
	}

	testCases := []struct {
		key                         []byte
		expectedLeft, expectedRight KVRange
	}{
		{
			[]byte{0x00},
			KVRange{},
			KVRange{
				NewKVPair([]byte{0x00}, []byte{0x00}),
				NewKVPair([]byte{0x01}, []byte{0x01}),
				NewKVPair([]byte{0x02}, []byte{0x02}),
			},
		},
		{
			[]byte{0x01},
			KVRange{
				NewKVPair([]byte{0x00}, []byte{0x00}),
			},
			KVRange{
				NewKVPair([]byte{0x01}, []byte{0x01}),
				NewKVPair([]byte{0x02}, []byte{0x02}),
			},
		},
		{
			[]byte{0x02},
			KVRange{
				NewKVPair([]byte{0x00}, []byte{0x00}),
				NewKVPair([]byte{0x01}, []byte{0x01}),
			},
			KVRange{
				NewKVPair([]byte{0x02}, []byte{0x02}),
			},
		},
		{
			[]byte{0x03},
			KVRange{
				NewKVPair([]byte{0x00}, []byte{0x00}),
				NewKVPair([]byte{0x01}, []byte{0x01}),
				NewKVPair([]byte{0x02}, []byte{0x02}),
			},
			KVRange{},
		},
	}

	for i, c := range testCases {
		left, right := kvrange.Split(c.key)
		require.Equal(t, c.expectedLeft, left, "Error spliting test: %d, value: %x", i, c.key)
		require.Equal(t, c.expectedRight, right, "Error spliting test: %d, value: %x", i, c.key)
	}
}

func TestGet(t *testing.T) {

	kvrange := NewKVRange()
	for i := 0; i < 10; i++ {
		kvpair := NewKVPair([]byte{byte(i)}, []byte{byte(i)})
		kvrange = kvrange.InsertSorted(kvpair)
	}

	tests := []struct {
		key, expectedValue []byte
	}{
		{[]byte{0x4}, []byte{0x4}},
	}

	for i, test := range tests {
		kvpair := kvrange.Get(test.key)
		require.Equalf(t, test.expectedValue, kvpair.Value, "Get error in test: %d, value: %x", i, test.key)
	}
}
