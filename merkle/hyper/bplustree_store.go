package hyper

import (
	"bytes"

	"github.com/cznic/b"
)

type BPlusTreeStorage struct {
	store *b.Tree
}

func NewBPlusTreeStorage() *BPlusTreeStorage {
	return &BPlusTreeStorage{b.TreeNew(cmp)}
}

func cmp(a, b interface{}) int {
	return bytes.Compare(a.([]byte), b.([]byte))
}

func (s *BPlusTreeStorage) Add(key []byte, value []byte) error {
	s.store.Set(key, value)
	return nil
}

func (s *BPlusTreeStorage) Get(p *Position) D {
	var d D
	var err error
	var k interface{}

	iter, _ := s.store.Seek(p.base)
	defer iter.Close()

	n := 0
	for {
		k, _, err = iter.Next()
		if err != nil {
			return d
		}
		if bytes.Compare(k.([]byte), p.split) < 0 {
			d = append(d, k.([]byte))
		} else {
			return d
		}
		n++
	}

}
