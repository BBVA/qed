package history

import (
	"math"

	"github.com/bbva/qed/balloon/common"
)

type HistoryTreeNavigator struct {
	version uint64
	depth   uint16
}

func NewHistoryTreeNavigator(version uint64) *HistoryTreeNavigator {
	depth := uint16(uint64(math.Ceil(math.Log2(float64(version + 1)))))
	return &HistoryTreeNavigator{version, depth}
}

func (n HistoryTreeNavigator) Root() common.Position {
	return NewPosition(0, n.depth)
}

func (n HistoryTreeNavigator) IsLeaf(pos common.Position) bool {
	return pos.Height() == 0
}

func (n HistoryTreeNavigator) IsRoot(pos common.Position) bool {
	return pos.Height() == n.depth && pos.IndexAsUint64() == 0
}

func (n HistoryTreeNavigator) GoToLeft(pos common.Position) common.Position {
	if pos.Height() == 0 {
		return nil
	}
	return NewPosition(pos.IndexAsUint64(), pos.Height()-1)
}
func (n HistoryTreeNavigator) GoToRight(pos common.Position) common.Position {
	rightIndex := pos.IndexAsUint64() + pow(2, pos.Height()-1)
	if pos.Height() == 0 || rightIndex > n.version {
		return nil
	}
	return NewPosition(rightIndex, pos.Height()-1)
}

func (n HistoryTreeNavigator) DescendToFirst(pos common.Position) common.Position {
	if n.IsLeaf(pos) {
		return nil
	}
	return NewPosition(pos.IndexAsUint64(), 0)
}

func (n HistoryTreeNavigator) DescendToLast(pos common.Position) common.Position {
	if n.IsLeaf(pos) {
		return nil
	}
	lastDescendantIndex := pos.IndexAsUint64() + pow(2, pos.Height()) - 1
	if lastDescendantIndex > n.version {
		return nil
	}
	return NewPosition(lastDescendantIndex, 0)
}

func pow(x, y uint16) uint64 {
	return uint64(math.Pow(float64(x), float64(y)))
}
