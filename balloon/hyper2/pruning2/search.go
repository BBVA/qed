package pruning2

import (
	"bytes"

	"github.com/bbva/qed/balloon/hyper2/navigation"
)

func PruneToFind(index []byte, batches BatchLoader) *OperationsStack {

	var traverse, traverseBatch, discardBranch func(pos navigation.Position, batch *BatchNode, iBatch int8, ops *OperationsStack)

	traverse = func(pos navigation.Position, batch *BatchNode, iBatch int8, ops *OperationsStack) {

		if batch == nil {
			batch = batches.Load(pos)
		}

		traverseBatch(pos, batch, iBatch, ops)
	}

	discardBranch = func(pos navigation.Position, batch *BatchNode, iBatch int8, ops *OperationsStack) {
		if batch.HasElementAt(iBatch) {
			ops.Push(getProvidedHash(pos, iBatch, batch))
		} else {
			ops.Push(getDefaultHash(pos))
		}
	}

	traverseBatch = func(pos navigation.Position, batch *BatchNode, iBatch int8, ops *OperationsStack) {

		// We found a nil value. That means there is no previous node stored on the current
		// path so we stop traversing because the index does no exist in the tree.
		// We return a new shortcut without mutating.
		if !batch.HasElementAt(iBatch) {
			ops.Push(shortcutHash(pos, index, nil)) // TODO shall i return nothing?
			return
		}

		// at the end of the batch tree
		if iBatch > 0 && pos.Height%4 == 0 {
			traverse(pos, nil, 0, ops)
			ops.Push(getProvidedHash(pos, iBatch, batch))
			return
		}

		// on an internal node of the subtree

		// we found a shortcut leaf in our path
		if batch.HasElementAt(iBatch) && batch.HasLeafAt(iBatch) {
			key, value := batch.GetLeafKVAt(iBatch)
			if bytes.Equal(index, key) {
				// we found the searched index
				ops.Push(shortcutHash(pos, key, value))
				return
			}
			// we found another shortcut leaf on our path so the index
			// we are looking for has never been inserted in the tree
			ops.Push(shortcutHash(pos, key, nil))
			return
		}

		rightPos := pos.Right()
		leftPos := pos.Left()
		if bytes.Compare(index, rightPos.Index) < 0 { // go to left
			traverse(pos.Left(), batch, 2*iBatch+1, ops)
			discardBranch(rightPos, batch, 2*iBatch+2, ops)
			ops.Push(collectHash(rightPos))
		} else { // go to right
			discardBranch(leftPos, batch, 2*iBatch+1, ops)
			ops.Push(collectHash(leftPos))
			traverse(rightPos, batch, 2*iBatch+2, ops)
		}

		ops.Push(innerHash(pos))

	}

	ops := NewOperationsStack()
	traverse(navigation.NewRootPosition(uint16(len(index)*8)), nil, 0, ops)
	return ops
}
