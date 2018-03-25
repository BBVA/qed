package hyper

import (
	bolt "github.com/coreos/bbolt"
)

type BoltStorage struct {


}

func (b *BoltStorage) Add(v *Value) error {
	b.ReplaceOrInsert(v)
	return nil
}

func (b *BoltStorage) Get(p *Position) *D {
	var d D
	var err error
	var k, v interface{}
	
	d.v = make([]*value, 0)
	
	b.AscendGreaterOrEqual(p.base, func(i Item) bool {
		if bytes.Compare(k.([]byte), p.split) < 0 {
			d.v = append(d.v, &value{k.([]byte), v.([]byte)})
			return true
		} 
		return false
	})

	return &d
}
