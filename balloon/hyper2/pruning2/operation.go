package pruning2

import (
	"github.com/bbva/qed/balloon/cache"
	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/hashing"
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
			c.AuditPath = append(c.AuditPath, hash)
			return hash
		},
	}
}
