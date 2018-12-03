package gossip

import (
	"bytes"

	"github.com/bbva/qed/protocol"

	"github.com/bbva/qed/util"
	"github.com/google/btree"
)

type LocalStore interface {
	Put(version uint64, snapshot protocol.Snapshot) error
	GetRange(start, end uint64) ([]protocol.Snapshot, error)
	DeleteRange(start, end uint64) error
}

type BPlusTreeStore struct {
	db *btree.BTree
}

type StoreItem struct {
	Key, Value []byte
}

func (p StoreItem) Less(b btree.Item) bool {
	return bytes.Compare(p.Key, b.(StoreItem).Key) < 0
}

func (s *BPlusTreeStore) Put(version uint64, snapshot protocol.Snapshot) error {
	encoded, err := snapshot.Encode()
	if err != nil {
		return err
	}
	s.db.ReplaceOrInsert(StoreItem{util.Uint64AsBytes(version), encoded})
	return nil
}

func (s BPlusTreeStore) GetRange(start, end uint64) ([]protocol.Snapshot, error) {
	result := make([]protocol.Snapshot, 0)
	startKey := util.Uint64AsBytes(start)
	endKey := util.Uint64AsBytes(end)
	s.db.AscendGreaterOrEqual(StoreItem{startKey, nil}, func(i btree.Item) bool {
		key := i.(StoreItem).Key
		if bytes.Compare(key, endKey) > 0 {
			return false
		}
		var snapshot protocol.Snapshot
		if err := snapshot.Decode(i.(StoreItem).Value); err != nil {
			return false
		}
		result = append(result, snapshot)
		return true
	})
	return result, nil
}

func (s *BPlusTreeStore) DeleteRange(start, end uint64) error {
	startKey := util.Uint64AsBytes(start)
	endKey := util.Uint64AsBytes(end)
	s.db.AscendGreaterOrEqual(StoreItem{startKey, nil}, func(i btree.Item) bool {
		key := i.(StoreItem).Key
		if bytes.Compare(key, endKey) > 0 {
			return false
		}
		s.db.Delete(i)
		return true
	})
	return nil
}
