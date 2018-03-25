package hyper

import (
	"bytes"
	"github.com/google/btree"
)

type BtreeStorage struct {
	store	*btree.Btree
}

func (v Value) Less(than btree.Item) bool {
	return bytes.Compare(v.k, than.(value).key) < 0
}

func (b *BtreeStorage) Add(v *Value) error {
	b.ReplaceOrInsert(v)
	return nil
}

func (b *BtreeStorage) Get(p *Position) *D {
	var d D
	var err error
	
	d.v = make([]*Value, 0)
	
	b.AscendGreaterOrEqual(p.base, func(i Item) bool {
		if bytes.Compare(i.(*Value).Key, p.split) < 0 {
			d.v = append(d.v, &value{i.(*Value).Key, i.(*Value).Value})
			return true
		} 
		return false
	})

	return &d
}
