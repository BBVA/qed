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

package pruning

import (
	"bytes"

	"github.com/bbva/qed/balloon/hyper/navigation"
)

func PruneToRebuild(index, serializedBatch []byte, cacheHeightLimit uint16, batches BatchLoader) *OperationsStack {

	persistedBatch := ParseBatchNode(len(index), serializedBatch)

	var traverse, discardBranch func(pos navigation.Position, batch *BatchNode, iBatch int8, ops *OperationsStack)

	discardBranch = func(pos navigation.Position, batch *BatchNode, iBatch int8, ops *OperationsStack) {

		if batch.HasElementAt(iBatch) {
			ops.Push(getProvidedHash(pos, iBatch, batch))
		} else {
			ops.Push(getDefaultHash(pos))
		}
	}

	traverse = func(pos navigation.Position, batch *BatchNode, iBatch int8, ops *OperationsStack) {

		// we don't need to check the length of the leaves because we
		// always have to descend to the cache height limit
		if pos.Height == cacheHeightLimit {
			ops.PushAll(useHash(pos, persistedBatch.GetElementAt(0)), updateBatchNode(pos, iBatch, batch))
			return
		}

		if batch == nil {
			batch = batches.Load(pos)
		}

		// at the end of a batch tree
		if iBatch > 0 && pos.Height%4 == 0 {
			traverse(pos, nil, 0, ops)
			ops.Push(updateBatchNode(pos, iBatch, batch))
			return
		}

		rightPos := pos.Right()
		leftPos := pos.Left()
		if bytes.Compare(index, rightPos.Index) < 0 { // go to left
			traverse(pos.Left(), batch, 2*iBatch+1, ops)
			discardBranch(rightPos, batch, 2*iBatch+2, ops)
		} else { // go to right
			discardBranch(leftPos, batch, 2*iBatch+1, ops)
			traverse(rightPos, batch, 2*iBatch+2, ops)
		}

		ops.PushAll(innerHash(pos), updateBatchNode(pos, iBatch, batch))
		if iBatch == 0 { // it's the root of the batch tree
			ops.Push(putInCache(pos, batch))
		}

	}

	ops := NewOperationsStack()
	traverse(navigation.NewRootPosition(uint16(len(index))), nil, 0, ops)
	return ops

}
