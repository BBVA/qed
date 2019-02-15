package pruning

import (
	"bytes"

	"github.com/bbva/qed/balloon/hyper2/navigation"
)

func PruneToFind(index []byte, batches BatchLoader) (Operation, error) {

	var traverse, traverseBatch func(pos *navigation.Position, batch *BatchNode, iBatch int8) (Operation, error)
	var discardBranch func(pos *navigation.Position, batch *BatchNode, iBatch int8) Operation

	traverse = func(pos *navigation.Position, batch *BatchNode, iBatch int8) (Operation, error) {

		var err error
		if batch == nil {
			batch, err = batches.Load(pos)
			if err != nil {
				return nil, err
			}
		}

		return traverseBatch(pos, batch, iBatch)
	}

	discardBranch = func(pos *navigation.Position, batch *BatchNode, iBatch int8) Operation {
		if batch.HasElementAt(iBatch) {
			return NewUseProvidedOp(pos, batch, iBatch)
		}
		return NewGetDefaultOp(pos)
	}

	traverseBatch = func(pos *navigation.Position, batch *BatchNode, iBatch int8) (Operation, error) {

		// We found a nil value. That means there is no previous node stored on the current
		// path so we stop traversing because the index does no exist in the tree.
		// We return a new shortcut without mutating.
		if !batch.HasElementAt(iBatch) {
			return NewShortcutLeafOp(pos, batch, iBatch, index, nil), nil
		}

		// at the end of the batch tree
		if iBatch > 0 && pos.Height%4 == 0 {
			op, err := traverse(pos, nil, 0)
			if err != nil {
				return nil, err
			}
			return NewLeafOp(pos, batch, iBatch, op), nil
		}

		// on an internal node of the subtree

		// we found a shortcut leaf in our path
		if batch.HasElementAt(iBatch) && batch.HasLeafAt(iBatch) {
			key, value := batch.GetLeafKVAt(iBatch)
			if bytes.Equal(index, key) {
				// we found the searched index
				return NewShortcutLeafOp(pos, batch, iBatch, key, value), nil
			}
			// we found another shortcut leaf on our path so the we index
			// we are looking for has never been inserted in the tree
			return NewShortcutLeafOp(pos, batch, iBatch, key, nil), nil
		}

		var left, right Operation
		var err error

		rightPos := pos.Right()
		leftPos := pos.Left()
		if bytes.Compare(index, rightPos.Index) < 0 { // go to left
			left, err = traverse(&leftPos, batch, 2*iBatch+1)
			if err != nil {
				return nil, err
			}
			right = NewCollectOp(discardBranch(&rightPos, batch, 2*iBatch+2))
		} else { // go to right
			left = NewCollectOp(discardBranch(&leftPos, batch, 2*iBatch+1))
			right, err = traverse(&rightPos, batch, 2*iBatch+2)
			if err != nil {
				return nil, err
			}
		}

		return NewInnerHashOp(pos, batch, iBatch, left, right), nil

	}

	root := navigation.NewRootPosition(uint16(len(index) * 8))
	return traverse(&root, nil, 0)
}
