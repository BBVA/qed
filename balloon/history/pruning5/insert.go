package pruning5

import (
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
)

type Traverse func(*navigation.Position) Operation

func PruneToInsert(version uint64, eventDigest hashing.Digest) Operation {

	var traverse Traverse
	traverse = func(pos *navigation.Position) Operation {

		if pos.IsLeaf() {
			return NewMutateOp(NewPutCacheOp(NewLeafHashOp(pos, eventDigest)))
		}

		var left, right Operation

		rightPos := pos.Right()
		if version < rightPos.Index { // go to left
			left = traverse(pos.Left())
			right = NewGetCacheOp(rightPos)
		} else { // go to right
			left = NewGetCacheOp(pos.Left())
			right = traverse(rightPos)
		}

		if rightPos.Index > version { // partial
			return NewPartialInnerHashOp(pos, left)
		}

		if pos.LastDescendant().Index <= version { // freeze
			return NewMutateOp(NewPutCacheOp(NewInnerHashOp(pos, left, right)))
		}
		return NewInnerHashOp(pos, left, right)

	}

	return traverse(navigation.NewRootPosition(version))
}
