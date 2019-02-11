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
	"github.com/bbva/qed/balloon/hyper2/navigation"
)

func pos(index byte, height uint16) *navigation.Position {
	return navigation.NewPosition([]byte{index}, height)
}

func inner(pos *navigation.Position, iBatch int8, batch []byte, left, right Operation) *InnerHashOp {
	return NewInnerHashOp(pos, ParseBatchNode(1, batch), iBatch, left, right)
}

func leaf(pos *navigation.Position, iBatch int8, batch []byte, op Operation) *LeafOp {
	return NewLeafOp(pos, ParseBatchNode(1, batch), iBatch, op)
}

func shortcut(pos *navigation.Position, iBatch int8, batch []byte, key, value []byte) *ShortcutLeafOp {
	return NewShortcutLeafOp(pos, ParseBatchNode(1, batch), iBatch, key, value)
}

func getDefault(pos *navigation.Position) *GetDefaultOp {
	return NewGetDefaultOp(pos)
}

func useProvided(pos *navigation.Position, iBatch int8, batch []byte) *UseProvidedOp {
	return NewUseProvidedOp(pos, ParseBatchNode(1, batch), iBatch)
}

func putBatch(op Operation, batch []byte) *PutBatchOp {
	return NewPutBatchOp(op, ParseBatchNode(1, batch))
}

func mutate(op Operation, batch []byte) *MutateBatchOp {
	return NewMutateBatchOp(op, ParseBatchNode(1, batch))
}

func collect(op Operation) *CollectOp {
	return NewCollectOp(op)
}

type FakeBatchLoader struct {
	cacheHeightLimit uint16
	cached           map[string]*BatchNode
	stored           map[string]*BatchNode
}

func NewFakeBatchLoader(cached map[string][]byte, stored map[string][]byte, cacheHeightLimit uint16) *FakeBatchLoader {
	loader := &FakeBatchLoader{
		cacheHeightLimit: cacheHeightLimit,
		cached:           make(map[string]*BatchNode, 0),
		stored:           make(map[string]*BatchNode, 0),
	}
	for k, v := range cached {
		loader.cached[k] = ParseBatchNode(1, v)
	}
	for k, v := range stored {
		loader.stored[k] = ParseBatchNode(1, v)
	}
	return loader
}

func (l *FakeBatchLoader) Load(pos *navigation.Position) (*BatchNode, error) {
	if pos.Height > l.cacheHeightLimit {
		batch, ok := l.cached[pos.StringId()]
		if ok {
			return batch, nil
		}
		return NewEmptyBatchNode(len(pos.Index)), nil
	}
	batch, ok := l.stored[pos.StringId()]
	if ok {
		return batch, nil
	}
	return NewEmptyBatchNode(len(pos.Index)), nil
}
