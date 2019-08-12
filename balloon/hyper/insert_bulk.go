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

package hyper

import (
	"github.com/bbva/qed/util"
)

func pruneToInsertBulk(indexes [][]byte, values [][]byte, cacheHeightLimit uint16, batches batchLoader) *operationsStack {

	var traverse, traverseThroughCache, traverseAfterCache traverseBatch

	traverse = func(pos position, leaves leavesList, batch *batchNode, iBatch int8, ops *operationsStack) {
		if batch == nil {
			batch = batches.Load(pos)
		}
		if pos.Height > cacheHeightLimit {
			traverseThroughCache(pos, leaves, batch, iBatch, ops)
		} else {
			traverseAfterCache(pos, leaves, batch, iBatch, ops)
		}
	}

	traverseThroughCache = func(pos position, leaves leavesList, batch *batchNode, iBatch int8, ops *operationsStack) {

		if len(leaves) == 0 { // discarded branch
			if batch.HasElementAt(iBatch) {
				ops.Push(getProvidedHash(pos, iBatch, batch))
			} else {
				ops.Push(getDefaultHash(pos))
			}
			return
		}

		// at the end of a batch tree
		if iBatch > 0 && pos.Height%4 == 0 {
			traverse(pos, leaves, nil, 0, ops)
			ops.Push(updateBatchNode(pos, iBatch, batch))
			return
		}

		// on an internal node with more than one leaf

		rightPos := pos.Right()
		leftLeaves, rightLeaves := leaves.Split(rightPos.Index)

		traverseThroughCache(pos.Left(), leftLeaves, batch, 2*iBatch+1, ops)
		traverseThroughCache(rightPos, rightLeaves, batch, 2*iBatch+2, ops)

		ops.PushAll(innerHash(pos), updateBatchNode(pos, iBatch, batch))
		if iBatch == 0 { // it's the root of the batch tree
			ops.Push(putInCache(pos, batch))
		}

	}

	traverseAfterCache = func(pos position, leaves leavesList, batch *batchNode, iBatch int8, ops *operationsStack) {

		if len(leaves) == 0 { // discarded branch
			if batch.HasElementAt(iBatch) {
				ops.Push(getProvidedHash(pos, iBatch, batch))
			} else {
				ops.Push(getDefaultHash(pos))
			}
			return
		}

		// at the end of the main tree
		// this is a special case because we have to mutate even if there exists a previous stored leaf (update scenario)
		if pos.IsLeaf() {
			if len(leaves) != 1 {
				panic("Oops, something went wrong. We cannot have more than one leaf at the end of the main tree")
			}
			// create or update the leaf with a new shortcut
			newBatch := newEmptyBatchNode(len(pos.Index))
			ops.PushAll(
				leafHash(pos, leaves[0].Value),
				updateBatchShortcut(pos, 0, newBatch, leaves[0].Index, leaves[0].Value),
				mutateBatch(pos, newBatch),
				updateBatchNode(pos, iBatch, batch),
			)
			return
		}

		// at the end of a subtree
		if iBatch > 0 && pos.Height%4 == 0 {
			if len(leaves) > 1 {
				// with more than one leaf to insert -> it's impossible to be a shortcut leaf
				traverse(pos, leaves, nil, 0, ops)
				ops.Push(updateBatchNode(pos, iBatch, batch))
				return
			}
			// with only one leaf to insert -> continue traversing
			if batch.HasElementAt(iBatch) {
				traverse(pos, leaves, nil, 0, ops)
				ops.Push(updateBatchNode(pos, iBatch, batch))
				return
			}
			// nil value (no previous node stored) so create a new shortcut batch
			newBatch := newEmptyBatchNode(len(pos.Index))
			ops.PushAll(
				leafHash(pos, leaves[0].Value),
				updateBatchShortcut(pos, 0, newBatch, leaves[0].Index, leaves[0].Value),
				mutateBatch(pos, newBatch),
				updateBatchNode(pos, iBatch, batch),
			)
			return
		}

		// on an internal node with only one leaf to insert

		if len(leaves) == 1 {
			// we found a nil in our path -> create a shortcut leaf
			if !batch.HasElementAt(iBatch) {
				ops.PushAll(
					leafHash(pos, leaves[0].Value),
					updateBatchShortcut(pos, iBatch, batch, leaves[0].Index, leaves[0].Value),
				)
				if pos.Height%4 == 0 { // at the root or at a leaf of the subtree (not necessary to check iBatch)
					ops.Push(mutateBatch(pos, batch))
				}
				return
			}
		}

		// on an internal node with more than one leaf

		// we found a node in our path and it is a shortcut leaf
		if batch.HasLeafAt(iBatch) {
			// push down leaf
			key, value := batch.GetLeafKVAt(iBatch)
			// we need to make a copy of the slice because it could affect to other branches
			newLeaves := append(leaves[:0:0], leaves...)
			newLeaves = newLeaves.InsertSorted(leaf{key, value})
			batch.ResetElementAt(iBatch)
			batch.ResetElementAt(2*iBatch + 1)
			batch.ResetElementAt(2*iBatch + 2)
			traverseAfterCache(pos, newLeaves, batch, iBatch, ops)
			return
		}

		rightPos := pos.Right()
		leftLeaves, rightLeaves := leaves.Split(rightPos.Index)

		traverseAfterCache(pos.Left(), leftLeaves, batch, 2*iBatch+1, ops)
		traverseAfterCache(rightPos, rightLeaves, batch, 2*iBatch+2, ops)

		ops.PushAll(innerHash(pos), updateBatchNode(pos, iBatch, batch))
		if iBatch == 0 { // at root node -> mutate batch
			ops.Push(mutateBatch(pos, batch))
		}

	}

	ops := newOperationsStack()
	leaves := make(leavesList, 0)
	indexLength := len(indexes[0])
	for i, index := range indexes {
		version := util.AddPaddingToBytes(values[i], indexLength)
		version = version[len(version)-indexLength:] // TODO GET RID OF THIS: used only to pass tests
		leaves = leaves.InsertSorted(leaf{index, version})
	}

	traverse(newRootPosition(uint16(indexLength)), leaves, nil, 0, ops)
	return ops
}
