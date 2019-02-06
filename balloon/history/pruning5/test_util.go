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

package pruning5

import (
	"github.com/bbva/qed/balloon/history/navigation"
)

func pos(index uint64, height uint16) *navigation.Position {
	return navigation.NewPosition(index, height)
}

func inner(pos *navigation.Position, left, right Operation) *InnerHashOp {
	return NewInnerHashOp(pos, left, right)
}

func partial(pos *navigation.Position, left Operation) *PartialInnerHashOp {
	return NewPartialInnerHashOp(pos, left)
}

func leaf(pos *navigation.Position, value byte) *LeafHashOp {
	return NewLeafHashOp(pos, []byte{value})
}

func leafnil(pos *navigation.Position) *LeafHashOp {
	return NewLeafHashOp(pos, nil)
}

func getCache(pos *navigation.Position) *GetCacheOp {
	return NewGetCacheOp(pos)
}

func putCache(op Operation) *PutCacheOp {
	return NewPutCacheOp(op)
}

func mutate(op Operation) *MutateOp {
	return NewMutateOp(op)
}

func collect(op Operation) *CollectOp {
	return NewCollectOp(op)
}
