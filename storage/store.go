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

package storage

import (
	"bytes"
	"errors"
	"io"
	"sort"
)

const (
	IndexPrefix        = byte(0x0)
	HyperCachePrefix   = byte(0x1)
	HistoryCachePrefix = byte(0x2)
	FSMStatePrefix     = byte(0x3)
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Store interface {
	Mutate(mutations []*Mutation) error
	GetRange(prefix byte, start, end []byte) (KVRange, error)
	Get(prefix byte, key []byte) (*KVPair, error)
	GetAll(prefix byte) KVPairReader
	GetLast(prefix byte) (*KVPair, error)
	Close() error
}

type DeletableStore interface {
	Store
	Delete(prefix byte, key []byte) error
}
type ManagedStore interface {
	Store
	Backup(w io.Writer, since uint64) error
	Load(r io.Reader) error
}

type Mutation struct {
	Prefix     byte
	Key, Value []byte
}

func NewMutation(prefix byte, key, value []byte) *Mutation {
	return &Mutation{prefix, key, value}
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

func (r KVRange) InsertSorted(p KVPair) KVRange {

	if len(r) == 0 {
		r = append(r, p)
		return r
	}

	index := sort.Search(len(r), func(i int) bool {
		return bytes.Compare(r[i].Key, p.Key) > 0
	})

	if index > 0 && bytes.Equal(r[index-1].Key, p.Key) {
		return r
	}

	r = append(r, p)
	copy(r[index+1:], r[index:])
	r[index] = p
	return r
}

func (r KVRange) Split(key []byte) (left, right KVRange) {
	// the smallest index i where r[i] >= index
	index := sort.Search(len(r), func(i int) bool {
		return bytes.Compare(r[i].Key, key) >= 0
	})
	return r[:index], r[index:]
}

func (r KVRange) Get(key []byte) KVPair {
	index := sort.Search(len(r), func(i int) bool {
		return bytes.Compare(r[i].Key, key) >= 0
	})
	if index < len(r) && bytes.Equal(r[index].Key, key) {
		return r[index]
	} else {
		panic("This should never happen")
	}
}
