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
	"fmt"

	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/storage"
)

type pruningContext struct {
	Hasher         hashing.Hasher
	Cache          cache.ModifiableCache
	RecoveryHeight uint16
	DefaultHashes  []hashing.Digest
	Mutations      []*storage.Mutation
	AuditPath      AuditPath
	Value          []byte
}

type operationCode int

const (
	leafHashCode operationCode = iota
	innerHashCode
	updateBatchNodeCode
	updateBatchShortcutCode
	getDefaultHashCode
	getProvidedHashCode
	putInCacheCode
	mutateBatchCode
	collectValueCode
	collectHashCode
	getFromPathCode
	noOpCode
)

type interpreter func(ops *operationsStack, c *pruningContext) (hashing.Digest, error)

type operation struct {
	Code      operationCode
	Pos       position
	Interpret interpreter
}

func leafHash(pos position, value []byte) *operation {
	return &operation{
		Code: leafHashCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			return c.Hasher.Salted(pos.Bytes(), value), nil
		},
	}
}

func innerHash(pos position) *operation {
	return &operation{
		Code: innerHashCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			leftHash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			rightHash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			return c.Hasher.Salted(pos.Bytes(), leftHash, rightHash), nil
		},
	}
}

func updateBatchNode(pos position, idx int8, batch *batchNode) *operation {
	return &operation{
		Code: updateBatchNodeCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			hash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			batch.AddHashAt(idx, hash)
			return hash, nil
		},
	}
}

func updateBatchShortcut(pos position, idx int8, batch *batchNode, key, value []byte) *operation {
	return &operation{
		Code: updateBatchShortcutCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			hash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			batch.AddLeafAt(idx, hash, key, value)
			return hash, nil
		},
	}
}

func getDefaultHash(pos position) *operation {
	return &operation{
		Code: getDefaultHashCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			return c.DefaultHashes[pos.Height], nil
		},
	}
}

func getProvidedHash(pos position, idx int8, batch *batchNode) *operation {
	return &operation{
		Code: getProvidedHashCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			return batch.GetElementAt(idx), nil
		},
	}
}

func putInCache(pos position, batch *batchNode) *operation {
	return &operation{
		Code: putInCacheCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {

			hash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			key := pos.Bytes()
			val := batch.Serialize()
			c.Cache.Put(key, val)
			if pos.Height == c.RecoveryHeight {
				c.Mutations = append(c.Mutations, storage.NewMutation(storage.HyperCacheTable, key, val))
			}
			return hash, nil
		},
	}
}

func mutateBatch(pos position, batch *batchNode) *operation {
	return &operation{
		Code: mutateBatchCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			hash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			c.Mutations = append(c.Mutations, storage.NewMutation(storage.HyperTable, pos.Bytes(), batch.Serialize()))
			return hash, nil
		},
	}
}

func collectValue(pos position, value []byte) *operation {
	return &operation{
		Code: collectValueCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			hash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			c.Value = value
			return hash, nil
		},
	}
}

func collectHash(pos position) *operation {
	return &operation{
		Code: collectHashCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			hash, err := ops.Pop().Interpret(ops, c)
			if err != nil {
				return nil, err
			}
			c.AuditPath[pos.StringId()] = hash
			return hash, nil
		},
	}
}

func getFromPath(pos position) *operation {
	return &operation{
		Code: getFromPathCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			hash, ok := c.AuditPath.Get(pos)
			if !ok {
				return nil, fmt.Errorf("Oops, something went wrong. Invalid position [%v] in audit path", pos)
			}
			return hash, nil
		},
	}
}

func noOp(pos position) *operation {
	return &operation{
		Code: noOpCode,
		Pos:  pos,
		Interpret: func(ops *operationsStack, c *pruningContext) (hashing.Digest, error) {
			return nil, nil
		},
	}
}

type operationsStack []*operation

func newOperationsStack() *operationsStack {
	return new(operationsStack)
}

func (s *operationsStack) Len() int {
	return len(*s)
}

func (s operationsStack) Peek() (op *operation) {
	return s[len(s)-1]
}

func (s *operationsStack) Pop() (op *operation) {
	i := s.Len() - 1
	op = (*s)[i]
	*s = (*s)[:i]
	return
}

func (s *operationsStack) Push(op *operation) {
	*s = append(*s, op)
}

func (s *operationsStack) PushAll(ops ...*operation) {
	*s = append(*s, ops...)
}

func (s *operationsStack) List() []*operation {
	l := make([]*operation, 0)
	for s.Len() > 0 {
		l = append(l, s.Pop())
	}
	return l
}
