package balloon

import (
	"encoding/binary"
	"fmt"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/storage"
)

type Balloon struct {
	history chan interface{}
	hyper   chan interface{}
	hasher  hashing.Hasher
	version uint
}

type Commitment struct {
	HistoryDigest []byte
	IndexDigest   []byte
	Version       uint
}

func NewBalloon(path string, cacheSize int, hasher hashing.Hasher) *Balloon {

	frozen := storage.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", path))
	leaves := storage.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", path))
	cache := storage.NewSimpleCache(cacheSize)

	history := history.NewTree(frozen, hasher)
	hyper := hyper.NewTree(path, cache, leaves, hasher)

	b := Balloon{
		make(chan interface{}),
		make(chan interface{}),
		hasher,
		0,
	}

	history.Run(b.history)
	hyper.Run(b.hyper)

	return &b

}

func (b *Balloon) Add(event []byte) (*Commitment, error) {
	digest := b.hasher(event)
	b.version++
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, uint64(b.version))

	historyAddResult:= make(chan []byte)
	hyperAddResult := make(chan []byte)

	b.history <- history.NewAdd(digest, index, historyAddResult)
	b.hyper <-hyper.NewAdd(digest, index, hyperAddResult)

	historyDigest := <-historyAddResult
	hyperDigest := <-hyperAddResult

	return &Commitment{historyDigest, hyperDigest, b.version}, nil
}

func (b *Balloon) Close() error {
	var result chan bool
	b.history <- history.NewStop(result)
	b.hyper <- hyper.NewStop(result)

	return nil
}
