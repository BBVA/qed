package pruning

import (
	"bytes"

	"github.com/bbva/qed/balloon/hyper2/navigation"
)

func PruneToVerify(key, value []byte, auditPathHeight uint16) Operation {

	var traverse func(pos *navigation.Position) Operation

	traverse = func(pos *navigation.Position) Operation {

		if pos.Height <= auditPathHeight {
			// we are at the leaf height
			return NewShortcutLeafOp(pos, nil, 0, key, value)
		}

		var left, right Operation
		rightPos := pos.Right()
		leftPos := pos.Left()
		if bytes.Compare(key, rightPos.Index) < 0 { // go to left
			left = traverse(&leftPos)
			right = NewGetDefaultOp(&rightPos)
		} else { // go to right
			left = NewGetDefaultOp(&leftPos)
			right = traverse(&rightPos)
		}

		return NewInnerHashOp(pos, nil, 0, left, right)

	}

	root := navigation.NewRootPosition(uint16(len(key) * 8))
	return traverse(&root)

}
