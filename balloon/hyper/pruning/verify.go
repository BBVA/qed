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

func PruneToVerify(index, value []byte, auditPathHeight uint16) *OperationsStack {

	var traverse func(pos navigation.Position, ops *OperationsStack)

	traverse = func(pos navigation.Position, ops *OperationsStack) {

		if pos.Height <= auditPathHeight {
			ops.Push(leafHash(pos, value))
			return
		}

		rightPos := pos.Right()
		if bytes.Compare(index, rightPos.Index) < 0 { // go to left
			traverse(pos.Left(), ops)
			ops.Push(getFromPath(rightPos))
		} else { // go to right
			ops.Push(getFromPath(pos.Left()))
			traverse(rightPos, ops)
		}

		ops.Push(innerHash(pos))

	}

	ops := NewOperationsStack()
	traverse(navigation.NewRootPosition(uint16(len(index))), ops)
	return ops

}
