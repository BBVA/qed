/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

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
	"github.com/bbva/qed/crypto/hashing"
)

func pruneToVerify(index, version uint64, eventDigest hashing.Digest) operation {

	var traverse func(pos *position) operation
	traverse = func(pos *position) operation {

		if pos.IsLeaf() {
			return newLeafHashOp(pos, eventDigest)
		}

		var left, right operation

		rightPos := pos.Right()
		if index < rightPos.Index { // go to left
			left = traverse(pos.Left())
			right = newGetCacheOp(rightPos)
		} else { // go to right
			left = newGetCacheOp(pos.Left())
			right = traverse(rightPos)
		}

		if rightPos.Index > version { // partial
			return newPartialInnerHashOp(pos, left)
		}

		return newInnerHashOp(pos, left, right)

	}

	return traverse(newRootPosition(version))
}

func pruneToVerifyIncrementalStart(version uint64) operation {

	var traverse func(pos *position) operation
	traverse = func(pos *position) operation {

		if pos.IsLeaf() {
			return newGetCacheOp(pos)
		}

		var left, right operation

		rightPos := pos.Right()
		if version < rightPos.Index { // go to left
			left = traverse(pos.Left())
			right = newGetCacheOp(rightPos)
		} else { // go to right
			left = newGetCacheOp(pos.Left())
			right = traverse(rightPos)
		}

		if rightPos.Index > version { // partial
			return newPartialInnerHashOp(pos, left)
		}
		return newInnerHashOp(pos, left, right)

	}

	return traverse(newRootPosition(version))
}

func pruneToVerifyIncrementalEnd(start, end uint64) operation {

	var traverse func(pos *position, targets targetsList) operation
	traverse = func(pos *position, targets targetsList) operation {

		if len(targets) == 0 {
			return newGetCacheOp(pos)
		}

		if pos.IsLeaf() {
			return newGetCacheOp(pos)
		}

		rightPos := pos.Right()
		leftTargets, rightTargets := targets.Split(rightPos.Index)

		left := traverse(pos.Left(), leftTargets)

		if end < rightPos.Index {
			return newPartialInnerHashOp(pos, left)
		}

		right := traverse(rightPos, rightTargets)

		return newInnerHashOp(pos, left, right)

	}

	targets := make(targetsList, 0)
	targets = targets.InsertSorted(start)
	targets = targets.InsertSorted(end)
	return traverse(newRootPosition(end), targets)
}
