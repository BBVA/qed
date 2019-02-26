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

func pruneToFind(version uint64) operation {

	var traverse func(pos *position) operation
	traverse = func(pos *position) operation {

		if pos.IsLeaf() {
			return newLeafHashOp(pos, nil)
		}

		var left, right operation

		rightPos := pos.Right()
		if version < rightPos.Index { // go to left
			left = traverse(pos.Left())
			right = newCollectOp(newGetCacheOp(rightPos))
		} else { // go to right
			left = newCollectOp(newGetCacheOp(pos.Left()))
			right = traverse(rightPos)
		}

		if rightPos.Index > version { // partial
			return newPartialInnerHashOp(pos, left)
		}
		return newInnerHashOp(pos, left, right)

	}

	return traverse(newRootPosition(version))
}
