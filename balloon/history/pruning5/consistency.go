package pruning5

import (
	"sort"

	"github.com/bbva/qed/balloon/history/navigation"
)

type Targets []uint64

func (t Targets) InsertSorted(version uint64) Targets {

	if len(t) == 0 {
		t = append(t, version)
		return t
	}

	index := sort.Search(len(t), func(i int) bool {
		return t[i] > version
	})

	if index > 0 && t[index-1] == version {
		return t
	}

	t = append(t, version)
	copy(t[index+1:], t[index:])
	t[index] = version
	return t

}

func (t Targets) Split(version uint64) (left, right Targets) {
	// the smallest index i where t[i] >= version
	index := sort.Search(len(t), func(i int) bool {
		return t[i] >= version //bytes.Compare(r[i].Key, key) >= 0
	})
	return t[:index], t[index:]
}

func PruneToFindConsistent(index, version uint64) Operation {

	var traverse func(pos *navigation.Position, targets Targets, shortcut bool) Operation

	traverse = func(pos *navigation.Position, targets Targets, shortcut bool) Operation {

		if len(targets) == 0 {
			if !shortcut {
				return NewCollectOp(NewGetCacheOp(pos))
			}
			return NewGetCacheOp(pos)
		}

		if pos.IsLeaf() {
			if pos.Index == index {
				return NewLeafHashOp(pos, nil)
			}
			if !shortcut {
				return NewCollectOp(NewGetCacheOp(pos))
			}
			return NewGetCacheOp(pos)
		}

		if len(targets) == 1 && targets[0] != index {
			if !shortcut {
				return NewCollectOp(traverse(pos, targets, true))
			}
		}

		rightPos := pos.Right()
		leftTargets, rightTargets := targets.Split(rightPos.Index)

		left := traverse(pos.Left(), leftTargets, shortcut)
		right := traverse(rightPos, rightTargets, shortcut)

		if version < rightPos.Index {
			return NewPartialInnerHashOp(pos, left)
		}

		return NewInnerHashOp(pos, left, right)
	}

	targets := make(Targets, 0)
	targets = targets.InsertSorted(index)
	targets = targets.InsertSorted(version)
	return traverse(navigation.NewRootPosition(version), targets, false)

}

func PruneToCheckConsistency(start, end uint64) Operation {

	var traverse func(pos *navigation.Position, targets Targets) Operation

	traverse = func(pos *navigation.Position, targets Targets) Operation {

		if len(targets) == 0 {
			return NewCollectOp(NewGetCacheOp(pos))
		}

		if pos.IsLeaf() {
			return NewCollectOp(NewGetCacheOp(pos))
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
