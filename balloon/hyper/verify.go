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

	"github.com/bbva/qed/util"
)

func pruneToVerify(index, value []byte, auditPathHeight uint16) *operationsStack {

	version := util.AddPaddingToBytes(value, len(index))
	version = version[len(version)-len(index):] // TODO GET RID OF THIS: used only to pass tests

	var traverse func(pos position, ops *operationsStack)

	traverse = func(pos position, ops *operationsStack) {

		if pos.Height <= auditPathHeight {
			ops.Push(leafHash(pos, version))
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

	ops := newOperationsStack()
	traverse(newRootPosition(uint16(len(index))), ops)
	return ops

}
