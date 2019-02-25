/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package history

import (
	"sort"
)

type targets []uint64

func (t targets) InsertSorted(version uint64) targets {

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

func (t targets) Split(version uint64) (left, right targets) {
	// the smallest index i where t[i] >= version
	index := sort.Search(len(t), func(i int) bool {
		return t[i] >= version //bytes.Compare(r[i].Key, key) >= 0
	})
	return t[:index], t[index:]
}

func pruneToFindConsistent(index, version uint64) operation {

	var traverse func(pos *position, targets targets, shortcut bool) operation

	traverse = func(pos *position, targets targets, shortcut bool) operation {

		if len(targets) == 0 {
			if !shortcut {
				return newCollectOp(newGetCacheOp(pos))
			}
			return newGetCacheOp(pos)
		}

		if pos.IsLeaf() {
			if pos.Index == index {
				return newLeafHashOp(pos, nil)
			}
			if !shortcut {
				return newCollectOp(newGetCacheOp(pos))
			}
			return newGetCacheOp(pos)
		}

		if len(targets) == 1 && targets[0] != index {
			if !shortcut {
				return newCollectOp(traverse(pos, targets, true))
			}
		}

		rightPos := pos.Right()
		leftTargets, rightTargets := targets.Split(rightPos.Index)

		left := traverse(pos.Left(), leftTargets, shortcut)
		right := traverse(rightPos, rightTargets, shortcut)

		if version < rightPos.Index {
			return newPartialInnerHashOp(pos, left)
		}

		return newInnerHashOp(pos, left, right)
	}

	targets := make(targets, 0)
	targets = targets.InsertSorted(index)
	targets = targets.InsertSorted(version)
	return traverse(newRootPosition(version), targets, false)

}

func pruneToCheckConsistency(start, end uint64) operation {

	var traverse func(pos *position, targets targets) operation

	traverse = func(pos *position, targets targets) operation {

		if len(targets) == 0 {
			return newCollectOp(newGetCacheOp(pos))
		}

		if pos.IsLeaf() {
			return newCollectOp(newGetCacheOp(pos))
		}

		rightPos := pos.Right()
		leftTargets, rightTargets := targets.Split(rightPos.Index)

		left := traverse(pos.Left(), leftTargets)
		right := traverse(rightPos, rightTargets)

		if end < rightPos.Index {
			return newPartialInnerHashOp(pos, left)
		}

		return newInnerHashOp(pos, left, right)
	}

	targets := make(targets, 0)
	targets = targets.InsertSorted(start)
	targets = targets.InsertSorted(end)
	return traverse(newRootPosition(end), targets)

}
