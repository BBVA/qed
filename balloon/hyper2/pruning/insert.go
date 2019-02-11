package pruning

import (
	"bytes"
	"sort"

	"github.com/bbva/qed/balloon/hyper2/navigation"
)

type Leaf struct {
	Index, Value []byte
}

type Leaves []Leaf

func (l Leaves) InsertSorted(leaf Leaf) Leaves {

	if len(l) == 0 {
		l = append(l, leaf)
		return l
	}

	index := sort.Search(len(l), func(i int) bool {
		return bytes.Compare(l[i].Index, leaf.Index) > 0
	})

	if index > 0 && bytes.Equal(l[index-1].Index, leaf.Index) {
		return l
	}

	l = append(l, leaf)
	copy(l[index+1:], l[index:])
	l[index] = leaf
	return l

}

func (l Leaves) Split(index []byte) (left, right Leaves) {
	// the smallest index i where l[i].Index >= index
	splitIndex := sort.Search(len(l), func(i int) bool {
		return bytes.Compare(l[i].Index, index) >= 0
	})
	return l[:splitIndex], l[splitIndex:]
}

type TraverseBatch func(pos *navigation.Position, leaves Leaves, batch *BatchNode, iBatch int8) (Operation, error)

func PruneToInsert(index []byte, value []byte, cacheHeightLimit uint16, batches BatchLoader) (Operation, error) {

	var traverse, traverseThroughCache, traverseAfterCache TraverseBatch

	traverse = func(pos *navigation.Position, leaves Leaves, batch *BatchNode, iBatch int8) (Operation, error) {

		var err error
		if batch == nil {
			batch, err = batches.Load(pos)
			if err != nil {
				return nil, err
			}
		}

		if pos.Height > cacheHeightLimit {
			return traverseThroughCache(pos, leaves, batch, iBatch)
		}
		return traverseAfterCache(pos, leaves, batch, iBatch)

	}

	traverseThroughCache = func(pos *navigation.Position, leaves Leaves, batch *BatchNode, iBatch int8) (Operation, error) {

		if len(leaves) == 0 { // discarded branch
			if batch.HasElementAt(iBatch) {
				return NewUseProvidedOp(pos, batch, iBatch), nil
			}
			return NewGetDefaultOp(pos), nil
		}

		// at the end of a batch tree
		if iBatch > 0 && pos.Height%4 == 0 {
			op, err := traverse(pos, leaves, nil, 0)
			if err != nil {
				return nil, err
			}
			return NewPutBatchOp(op, batch), nil
		}

		// on an internal node of the subtree

		// we found a node in our path
		if batch.HasElementAt(iBatch) {
			// we found a shortcut leaf in our path
			if batch.HasLeafAt(iBatch) {
				// push down leaf
				key, value := batch.GetLeafKVAt(iBatch)
				leaves = leaves.InsertSorted(Leaf{key, value})
				batch.ResetElementAt(iBatch)
				batch.ResetElementAt(2*iBatch + 1)
				batch.ResetElementAt(2*iBatch + 2)
				return traverse(pos, leaves, batch, iBatch)
			}
		}

		// on an internal node with more than one leaf

		rightPos := pos.Right()
		leftLeaves, rightLeaves := leaves.Split(rightPos.Index)

		left, err := traverse(pos.Left(), leftLeaves, batch, 2*iBatch+1)
		if err != nil {
			return nil, err
		}
		right, err := traverse(rightPos, rightLeaves, batch, 2*iBatch+2)
		if err != nil {
			return nil, err
		}

		op := NewInnerHashOp(pos, batch, iBatch, left, right)
		if iBatch == 0 { // it's the root of the batch tree
			return NewPutBatchOp(op, batch), nil
		}
		return op, nil

	}

	traverseAfterCache = func(pos *navigation.Position, leaves Leaves, batch *BatchNode, iBatch int8) (Operation, error) {

		if len(leaves) == 0 { // discarded branch
			if batch.HasElementAt(iBatch) {
				return NewUseProvidedOp(pos, batch, iBatch), nil
			}
			return NewGetDefaultOp(pos), nil
		}

		// at the end of the main tree
		// this is a special case because we have to mutate even if there exists a previous stored leaf (update scenario)
		if pos.IsLeaf() {
			if len(leaves) != 1 {
				panic("Oops, something went wrong. We cannot have more than one leaf at the end of the main tree")
			}
			// create or update the leaf with a new shortcut
			newBatch := NewEmptyBatchNode(len(pos.Index))
			shortcut := NewMutateBatchOp(
				NewShortcutLeafOp(pos, newBatch, 0, leaves[0].Index, leaves[0].Value),
				newBatch,
			)
			return NewLeafOp(pos, batch, iBatch, shortcut), nil
		}

		// at the end of a subtree
		if iBatch > 0 && pos.Height%4 == 0 {
			if len(leaves) > 1 {
				// with more than one leaf to insert -> it's impossible to be a shortcut leaf
				op, err := traverse(pos, leaves, nil, 0)
				if err != nil {
					return nil, err
				}
				return NewLeafOp(pos, batch, iBatch, op), nil
			}
			// with only one leaf to insert -> add a new shortcut leaf or continue traversing
			if batch.HasElementAt(iBatch) {
				// continue traversing
				op, err := traverse(pos, leaves, nil, 0)
				if err != nil {
					return nil, err
				}
				return NewLeafOp(pos, batch, iBatch, op), nil
			}
			// nil value (no previous node stored) so create a new shortcut batch
			newBatch := NewEmptyBatchNode(len(pos.Index))
			shortcut := NewMutateBatchOp(
				NewShortcutLeafOp(pos, newBatch, 0, leaves[0].Index, leaves[0].Value),
				newBatch,
			)
			return NewLeafOp(pos, batch, iBatch, shortcut), nil
		}

		// on an internal node with only one leaf to insert

		if len(leaves) == 1 {
			// we found a nil in our path -> create a shortcut leaf
			if !batch.HasElementAt(iBatch) {
				shortcut := NewShortcutLeafOp(pos, batch, iBatch, leaves[0].Index, leaves[0].Value)
				if pos.Height%4 == 0 { // at the root or at leaf of the subtree
					return NewMutateBatchOp(shortcut, NewEmptyBatchNode(len(pos.Index))), nil
				}
				return shortcut, nil
			}

			// we found a node in our path
			if batch.HasElementAt(iBatch) {
				// we found a shortcut leaf in our path
				if batch.HasLeafAt(iBatch) {
					// push down leaf
					key, value := batch.GetLeafKVAt(iBatch)
					leaves = leaves.InsertSorted(Leaf{key, value})
					batch.ResetElementAt(iBatch)
					batch.ResetElementAt(2*iBatch + 1)
					batch.ResetElementAt(2*iBatch + 2)
					return traverse(pos, leaves, batch, iBatch)
				}
			}
		}

		// on an internal node with more than one leaf
		rightPos := pos.Right()
		leftLeaves, rightLeaves := leaves.Split(rightPos.Index)

		left, err := traverse(pos.Left(), leftLeaves, batch, 2*iBatch+1)
		if err != nil {
			return nil, err
		}
		right, err := traverse(rightPos, rightLeaves, batch, 2*iBatch+2)
		if err != nil {
			return nil, err
		}

		op := NewInnerHashOp(pos, batch, iBatch, left, right)
		if iBatch == 0 { // at root node -> mutate batch
			return NewMutateBatchOp(op, batch), nil
		}
		return op, nil

	}

	leaves := make(Leaves, 0)
	leaves = leaves.InsertSorted(Leaf{index, value})
	return traverse(navigation.NewRootPosition(uint16(len(index)*8)), leaves, nil, 0)
}
