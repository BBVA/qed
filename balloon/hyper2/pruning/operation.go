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
	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
)

type Context struct {
	Hasher        hashing.Hasher
	Cache         cache.ModifiableCache
	DefaultHashes []hashing.Digest
	Mutations     []*storage.Mutation
	AuditPath     navigation.AuditPath
}

type OperationCode int

const (
	LeafHashCode OperationCode = iota
	InnerHashCode
	UpdateBatchNodeCode
	UpdateBatchShortcutCode
	GetDefaultHashCode
	GetProvidedHashCode
	PutInCacheCode
	MutateBatchCode
	CollectHashCode
	GetFromPathCode
	UseHashCode
	NoOpCode
)

type Interpreter func(ops *OperationsStack, c *Context) hashing.Digest

type Operation struct {
	Code      OperationCode
	Pos       navigation.Position
	Interpret Interpreter
}

func leafHash(pos navigation.Position, value []byte) *Operation {
	return &Operation{
		Code: LeafHashCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			return c.Hasher.Salted(pos.Bytes(), value)
		},
	}
}

func innerHash(pos navigation.Position) *Operation {
	return &Operation{
		Code: InnerHashCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			leftHash := ops.Pop().Interpret(ops, c)
			rightHash := ops.Pop().Interpret(ops, c)
			return c.Hasher.Salted(pos.Bytes(), leftHash, rightHash)
		},
	}
}

func updateBatchNode(pos navigation.Position, idx int8, batch *BatchNode) *Operation {
	return &Operation{
		Code: UpdateBatchNodeCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			hash := ops.Pop().Interpret(ops, c)
			batch.AddHashAt(idx, hash)
			return hash
		},
	}
}

func updateBatchShortcut(pos navigation.Position, idx int8, batch *BatchNode, key, value []byte) *Operation {
	return &Operation{
		Code: UpdateBatchShortcutCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			hash := ops.Pop().Interpret(ops, c)
			batch.AddLeafAt(idx, hash, key, value)
			return hash
		},
	}
}

func getDefaultHash(pos navigation.Position) *Operation {
	return &Operation{
		Code: GetDefaultHashCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			return c.DefaultHashes[pos.Height]
		},
	}
}

func getProvidedHash(pos navigation.Position, idx int8, batch *BatchNode) *Operation {
	return &Operation{
		Code: GetProvidedHashCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			return batch.GetElementAt(idx)
		},
	}
}

func putInCache(pos navigation.Position, batch *BatchNode) *Operation {
	return &Operation{
		Code: PutInCacheCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			hash := ops.Pop().Interpret(ops, c)
			c.Cache.Put(pos.Bytes(), batch.Serialize())
			return hash
		},
	}
}

func mutateBatch(pos navigation.Position, batch *BatchNode) *Operation {
	return &Operation{
		Code: MutateBatchCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			hash := ops.Pop().Interpret(ops, c)
			c.Mutations = append(c.Mutations, storage.NewMutation(storage.HyperCachePrefix, pos.Bytes(), batch.Serialize()))
			return hash
		},
	}
}

func collectHash(pos navigation.Position) *Operation {
	return &Operation{
		Code: CollectHashCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			hash := ops.Pop().Interpret(ops, c)
			c.AuditPath[pos.StringId()] = hash
			return hash
		},
	}
}

func getFromPath(pos navigation.Position) *Operation {
	return &Operation{
		Code: GetFromPathCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			hash, ok := c.AuditPath.Get(pos)
			if !ok {
				log.Fatalf("Oops, something went wrong. Invalid position in audit path")
			}
			return hash
		},
	}
}

func useHash(pos navigation.Position, digest []byte) *Operation {
	return &Operation{
		Code: UseHashCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			return digest
		},
	}
}

func noOp(pos navigation.Position) *Operation {
	return &Operation{
		Code: NoOpCode,
		Pos:  pos,
		Interpret: func(ops *OperationsStack, c *Context) hashing.Digest {
			return nil
		},
	}
}
