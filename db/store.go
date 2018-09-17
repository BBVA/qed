package db

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
)

var (
	ErrKeyNotFound = errors.New("key not found")
)

type Store interface {
	Mutate(mutations ...Mutation) error
	GetRange(prefix byte, start, end []byte) (KVRange, error)
	Get(prefix byte, key []byte) (*KVPair, error)
	GetAll(prefix byte) KVPairReader
	GetLast(prefix byte) (*KVPair, error)
	Close() error
}

type DeletableStore interface {
	Store
	Delete(key []byte) error
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

func getIndex(r KVRange, key []byte) int {
	return sort.Search(len(r), func(i int) bool {
		return bytes.Compare(r[i].Key, key) >= 0
	})
}

func (r KVRange) InsertSorted(p KVPair) KVRange {
	index := getIndex(r, p.Key)
	r = append(r, p)
	copy(r[index+1:], r[index:])
	r[index] = p
	return r
}

func (r KVRange) Split(key []byte) (left, right KVRange) {
	// the smallest index i where r[i] >= index
	index := getIndex(r, key)
	return r[:index], r[index:]
}

func (r KVRange) Get(key []byte) KVPair {
	index := getIndex(r, key)
	if index < len(r) && bytes.Equal(r[index].Key, key) {
		return r[index]
	} else {
		panic("This should never happen")
	}
}
