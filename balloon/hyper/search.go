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
	"bytes"
)

func pruneToFind(index []byte, batches batchLoader) *operationsStack {

	var traverse, traverseBatch, discardBranch func(pos position, batch *batchNode, iBatch int8, ops *operationsStack)

	traverse = func(pos position, batch *batchNode, iBatch int8, ops *operationsStack) {
		if batch == nil {
			batch = batches.Load(pos)
		}
		traverseBatch(pos, batch, iBatch, ops)
	}

	discardBranch = func(pos position, batch *batchNode, iBatch int8, ops *operationsStack) {
		if batch.HasElementAt(iBatch) {
			ops.PushAll(getProvidedHash(pos, iBatch, batch), collectHash(pos))
		} else {
			ops.PushAll(getDefaultHash(pos), collectHash(pos))
		}
	}

	traverseBatch = func(pos position, batch *batchNode, iBatch int8, ops *operationsStack) {

		// We found a nil value. That means there is no previous node stored on the current
		// path so we stop traversing because the index does no exist in the tree.
		if !batch.HasElementAt(iBatch) {
			ops.Push(noOp(pos))
			return
		}

		// at the end of the batch tree
		if iBatch > 0 && pos.Height%4 == 0 {
			traverse(pos, nil, 0, ops) // load another batch
			return
		}

		// on an internal node of the subtree

		// we found a shortcut leaf in our path
		if batch.HasLeafAt(iBatch) {
			// regardless if the key of the shortcut matches the searched index
			// we must stop traversing because there are no more leaves below
			ops.Push(getProvidedHash(pos, iBatch, batch)) // not collected
			k, v := batch.GetLeafKVAt(iBatch)
			if bytes.Equal(k, index) {
				ops.Push(collectValue(pos, v)) // collect value if the key matches the queried index
			}
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

		ops.Push(innerHash(pos))
	}

	ops := newOperationsStack()
	root := newRootPosition(uint16(len(index)))
	traverse(root, nil, 0, ops)
	if ops.Len() == 0 {
		ops.Push(noOp(root))
	}
	return ops
}
