package history

import (
	"testing"

	"github.com/bbva/qed/balloon2/common"
	"github.com/stretchr/testify/require"
)

func TestSingleTargetedCacheResolver(t *testing.T) {

	testCases := []struct {
		version  uint64
		position common.Position
		ok       bool
	}{
		{0, NewPosition(0, 0), false},

		{1, NewPosition(0, 0), true},
		{1, NewPosition(1, 0), false},
		{1, NewPosition(0, 1), false},

		{2, NewPosition(0, 0), true},
		{2, NewPosition(1, 0), true},
		{2, NewPosition(2, 0), false},
		{2, NewPosition(0, 1), true},
		{2, NewPosition(0, 2), false},

		{3, NewPosition(0, 0), true},
		{3, NewPosition(1, 0), true},
		{3, NewPosition(2, 0), true},
		{3, NewPosition(3, 0), false},
		{3, NewPosition(0, 1), true},
		{3, NewPosition(0, 2), false},
		{3, NewPosition(2, 1), false},

		{4, NewPosition(0, 0), true},
		{4, NewPosition(1, 0), true},
		{4, NewPosition(2, 0), true},
		{4, NewPosition(3, 0), true},
		{4, NewPosition(4, 0), false},
		{4, NewPosition(0, 1), true},
		{4, NewPosition(0, 2), true},
		{4, NewPosition(0, 3), false},
	}

	for i, c := range testCases {
		resolver := NewSingleTargetedCacheResolver(c.version)
		result := resolver.ShouldGetFromCache(c.position)
		require.Equalf(t, c.ok, result, "The result should match for test case %d", i)
	}

}

func TestNewDoubleTargetedCacheResolver(t *testing.T) {

	testCases := []struct {
		start, end uint64
		position   common.Position
		ok         bool
	}{
		{0, 0, NewPosition(0, 0), false},

		{0, 1, NewPosition(0, 0), false},
		{0, 1, NewPosition(1, 0), true},
		{0, 1, NewPosition(0, 1), false},

		{0, 2, NewPosition(0, 0), false},
		{0, 2, NewPosition(1, 0), true},
		{0, 2, NewPosition(0, 1), false},
		{0, 2, NewPosition(2, 0), true},

		{0, 3, NewPosition(0, 0), false},
		{0, 3, NewPosition(1, 0), true},
		{0, 3, NewPosition(0, 1), false},
		{0, 3, NewPosition(2, 0), true},
		{0, 3, NewPosition(3, 0), true},

		{0, 4, NewPosition(0, 0), false},
		{0, 4, NewPosition(1, 0), true},
		{0, 4, NewPosition(2, 1), true},
		{0, 4, NewPosition(4, 0), true},
	}

	for i, c := range testCases {
		resolver := NewDoubleTargetedCacheResolver(c.start, c.end)
		result := resolver.ShouldGetFromCache(c.position)
		require.Equalf(t, c.ok, result, "The result should match for test case %d", i)
	}

}

func TestIncrementalCacheResolver(t *testing.T) {

	testCases := []struct {
		start, end uint64
		position   common.Position
		ok         bool
	}{
		{0, 0, NewPosition(0, 0), true},

		{0, 1, NewPosition(0, 0), true},
		{0, 1, NewPosition(1, 0), true},
		{0, 1, NewPosition(0, 1), false},

		{0, 2, NewPosition(0, 0), true},
		{0, 2, NewPosition(1, 0), true},
		{0, 2, NewPosition(0, 1), false},
		{0, 2, NewPosition(2, 0), true},

		{0, 3, NewPosition(0, 0), true},
		{0, 3, NewPosition(1, 0), true},
		{0, 3, NewPosition(0, 1), false},
		{0, 3, NewPosition(2, 0), true},
		{0, 3, NewPosition(3, 0), true},

		{0, 4, NewPosition(0, 0), true},
		{0, 4, NewPosition(1, 0), true},
		{0, 4, NewPosition(2, 1), true},
		{0, 4, NewPosition(4, 0), true},
	}

	for i, c := range testCases {
		resolver := NewIncrementalCacheResolver(c.start, c.end)
		result := resolver.ShouldGetFromCache(c.position)
		require.Equalf(t, c.ok, result, "The result should match for test case %d", i)
	}

}
