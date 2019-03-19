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

package history

func pos(index uint64, height uint16) *position {
	return newPosition(index, height)
}

func inner(pos *position, left, right operation) *innerHashOp {
	return newInnerHashOp(pos, left, right)
}

func partial(pos *position, left operation) *partialInnerHashOp {
	return newPartialInnerHashOp(pos, left)
}

func leaf(pos *position, value byte) *leafHashOp {
	return newLeafHashOp(pos, []byte{value})
}

func leafnil(pos *position) *leafHashOp {
	return newLeafHashOp(pos, nil)
}

func getCache(pos *position) *getCacheOp {
	return newGetCacheOp(pos)
}

func putCache(op operation) *putCacheOp {
	return newPutCacheOp(op)
}

func mutate(op operation) *mutateOp {
	return newMutateOp(op)
}

func collect(op operation) *collectOp {
	return newCollectOp(op)
}
