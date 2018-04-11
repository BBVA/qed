package balloon

import (
	"encoding/binary"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/storage"
)

type Balloon struct {
	history *history.Tree
	hyper   *hyper.Tree
	hasher  hashing.Hasher
	version uint
	ops     chan interface{} // serialize operations
}

type Commitment struct {
	HistoryDigest []byte
	IndexDigest   []byte
	Version       uint
}

func NewBalloon(path string, cacheSize int, hasher hashing.Hasher, frozen, leaves storage.Store, cache storage.Cache) *Balloon {

	history := history.NewTree(frozen, hasher)
	hyper := hyper.NewTree(path, cache, leaves, hasher)

	b := Balloon{
		history,
		hyper,
		hasher,
		0,
		nil,
	}
	b.ops = b.operations()
	return &b

}

func (b *Balloon) add(event []byte) (*Commitment, error) {
	digest := b.hasher(event)
	b.version++
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, uint64(b.version))

	return &Commitment{
		<-b.history.Add(digest, index),
		<-b.hyper.Add(index, digest),
		b.version,
	}, nil
}

// Run listens in channel operations to execute in the tree
func (b *Balloon) operations() chan interface{} {
	operations := make(chan interface{}, 0)
	go func() {
		for {
			select {
			case op := <-operations:
				switch msg := op.(type) {
				case *close:
					msg.result <- true
					return
				case *add:
					digest, _ := b.add(msg.event)
					msg.result <- digest
				default:
					panic("Hyper tree Run() message not implemented!!")
				}

			}
		}
	}()
	return operations
}

type add struct {
	event  []byte
	result chan *Commitment
}

func (b Balloon) Add(event []byte) chan *Commitment {
	result := make(chan *Commitment)
	b.ops <- &add{
		event,
		result,
	}

	return result
}

type close struct {
	stop   bool
	result chan bool
}

func (b *Balloon) Close() chan bool {
	result := make(chan bool)

	b.history.Close()
	b.hyper.Close()

	b.ops <- &close{true, result}
	return result
}
