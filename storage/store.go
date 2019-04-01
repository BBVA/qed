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

package storage

import (
	"errors"
	"io"

	metrics "github.com/bbva/qed/metrics"
)

// Table groups related key-value pairs under a
// consistent space.
type Table uint32

const (
	// DefaultTable is mandatory but not used.
	DefaultTable Table = iota
	// IndexTable contains maps between event hashes and versions.
	// H(event) -> version
	IndexTable
	// HyperCacheTable contains cached batches of the hyper tree.
	// Position -> Batch
	HyperCacheTable
	// HistoryCacheTable contains cached hashes of the history tree.
	// Position -> Hash
	HistoryCacheTable
	// FSMStateTable contains the current state of the FSM (index, term, version...).
	// key -> state
	FSMStateTable
)

// FSMStateTableKey single key to persist fsm state.
var FSMStateTableKey = []byte{0xab}

// String returns a string representation of the table.
func (t Table) String() string {
	var s string
	switch t {
	case DefaultTable:
		s = "default"
	case IndexTable:
		s = "index"
	case HyperCacheTable:
		s = "hyper"
	case HistoryCacheTable:
		s = "history"
	case FSMStateTable:
		s = "fsm"
	}
	return s
}

// Prefix returns the byte prefix associated with this table.
// This method exists for backward compatibility purposes.
func (t Table) Prefix() byte {
	var prefix byte
	switch t {
	case IndexTable:
		prefix = byte(0x0)
	case HyperCacheTable:
		prefix = byte(0x1)
	case HistoryCacheTable:
		prefix = byte(0x2)
	case FSMStateTable:
		prefix = byte(0x3)
	default:
		prefix = byte(0x4)
	}
	return prefix
}

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Store interface {
	Mutate(mutations []*Mutation) error
	GetRange(table Table, start, end []byte) (KVRange, error)
	Get(table Table, key []byte) (*KVPair, error)
	GetAll(table Table) KVPairReader
	GetLast(table Table) (*KVPair, error)
	Close() error
}

type ManagedStore interface {
	Store
	Backup(w io.Writer, until uint64) error
	Snapshot() (uint64, error)
	Load(r io.Reader) error
	RegisterMetrics(metrics.Registry)
}

type Mutation struct {
	Table      Table
	Key, Value []byte
}

func NewMutation(table Table, key, value []byte) *Mutation {
	return &Mutation{
		Table: table,
		Key:   key,
		Value: value,
	}
}

type KVPair struct {
	Key, Value []byte
}

func NewKVPair(key, value []byte) KVPair {
	return KVPair{Key: key, Value: value}
}

type KVPairReader interface {
	Read([]*KVPair) (n int, err error)
	Close()
}

type KVRange []KVPair

func NewKVRange() KVRange {
	return make(KVRange, 0)
}
