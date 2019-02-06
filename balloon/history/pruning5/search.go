package pruning5

import (
	"github.com/bbva/qed/balloon/history/navigation"
)

func PruneToFind(version uint64) Operation {

	var traverse Traverse
	traverse = func(pos *navigation.Position) Operation {

		if pos.IsLeaf() {
			return NewLeafHashOp(pos, nil)
		}

		var left, right Operation

		rightPos := pos.Right()
		if version < rightPos.Index { // go to left
			left = traverse(pos.Left())
			right = NewCollectOp(NewGetCacheOp(rightPos))
		} else { // go to right
			left = NewCollectOp(NewGetCacheOp(pos.Left()))
			right = traverse(rightPos)
		}

		if rightPos.Index > version { // partial
			return NewPartialInnerHashOp(pos, left)
		}
		return NewInnerHashOp(pos, left, right)

	}

	return traverse(navigation.NewRootPosition(version))
}
