package pruning

import (
	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
)

func PruneToVerify(index, version uint64, eventDigest hashing.Digest) Operation {

	var traverse Traverse
	traverse = func(pos *navigation.Position) Operation {

		if pos.IsLeaf() {
			return NewLeafHashOp(pos, eventDigest)
		}

		var left, right Operation

		rightPos := pos.Right()
		if index < rightPos.Index { // go to left
			left = traverse(pos.Left())
			right = NewGetCacheOp(rightPos)
		} else { // go to right
			left = NewGetCacheOp(pos.Left())
			right = traverse(rightPos)
		}

		if rightPos.Index > version { // partial
			return NewPartialInnerHashOp(pos, left)
		}

		return NewInnerHashOp(pos, left, right)

	}

	return traverse(navigation.NewRootPosition(version))
}

func PruneToVerifyIncrementalStart(version uint64) Operation {

	var traverse Traverse
	traverse = func(pos *navigation.Position) Operation {

		if pos.IsLeaf() {
			return NewGetCacheOp(pos)
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
		return NewInnerHashOp(pos, left, right)

	}

	return traverse(navigation.NewRootPosition(version))
}

func PruneToVerifyIncrementalEnd(start, end uint64) Operation {

	var traverse func(pos *navigation.Position, targets Targets) Operation

	traverse = func(pos *navigation.Position, targets Targets) Operation {

		if len(targets) == 0 {
			return NewGetCacheOp(pos)
		}

		if pos.IsLeaf() {
			return NewGetCacheOp(pos)
		}

		rightPos := pos.Right()
		leftTargets, rightTargets := targets.Split(rightPos.Index)

		left := traverse(pos.Left(), leftTargets)
		right := traverse(rightPos, rightTargets)

		if end < rightPos.Index {
			return NewPartialInnerHashOp(pos, left)
		}

		return NewInnerHashOp(pos, left, right)

	}

	targets := make(Targets, 0)
	targets = targets.InsertSorted(start)
	targets = targets.InsertSorted(end)
	return traverse(navigation.NewRootPosition(end), targets)
}
