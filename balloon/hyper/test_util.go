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

func pos(index byte, height uint16) position {
	return newPosition([]byte{index}, height)
}

type op struct {
	Code operationCode
	Pos  position
}

type fakeBatchLoader struct {
	cacheHeightLimit uint16
	cached           map[string]*batchNode
	stored           map[string]*batchNode
}

func newFakeBatchLoader(cached map[string][]byte, stored map[string][]byte, cacheHeightLimit uint16) *fakeBatchLoader {
	loader := &fakeBatchLoader{
		cacheHeightLimit: cacheHeightLimit,
		cached:           make(map[string]*batchNode, 0),
		stored:           make(map[string]*batchNode, 0),
	}
	for k, v := range cached {
		loader.cached[k] = parseBatchNode(1, v)
	}
	for k, v := range stored {
		loader.stored[k] = parseBatchNode(1, v)
	}
	return loader
}

func (l *fakeBatchLoader) Load(pos position) *batchNode {
	if pos.Height > l.cacheHeightLimit {
		batch, ok := l.cached[pos.StringId()]
		if !ok {
			return newEmptyBatchNode(len(pos.Index))
		}
		return batch
	}
	batch, ok := l.stored[pos.StringId()]
	if !ok {
		return newEmptyBatchNode(len(pos.Index))
	}
	return batch
}
