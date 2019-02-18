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

package pruning2

import (
	"github.com/bbva/qed/balloon/hyper2/navigation"
)

func pos(index byte, height uint16) navigation.Position {
	return navigation.NewPosition([]byte{index}, height)
}

type op struct {
	Code OperationCode
	Pos  navigation.Position
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

func (l *FakeBatchLoader) Load(pos navigation.Position) *BatchNode {
	if pos.Height > l.cacheHeightLimit {
		batch, ok := l.cached[pos.StringId()]
		if !ok {
			return NewEmptyBatchNode(len(pos.Index))
		}
		return batch
	}
	batch, ok := l.stored[pos.StringId()]
	if !ok {
		return NewEmptyBatchNode(len(pos.Index))
	}
	return batch
}
