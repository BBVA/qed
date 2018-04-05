package balloon

import (
	"encoding/binary"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/storage"
)

type Balloon struct {
	history chan interface{}
	hyper   chan interface{}
	store   storage.Store
	hasher  hashing.Hasher
	version uint
}

type Commitment struct {
	HistoryDigest []byte
	IndexDigest   []byte
	Version       uint
}

func NewBalloon(bs, hs, hys storage.Store, hasher hashing.Hasher) *Balloon {

	hc := make(chan interface{})
	hyc := make(chan interface{})
	history := history.NewTree(hs, hasher)
	hyper := hyper.NewTree("id", hasher, 256, 30, storage.NewSimpleCache(50000000), hys)

	history.Run(hc)
	hyper.Run(hyc)

	return &Balloon{
		hc,
		hyc,
		bs,
		hasher,
		0,
	}
}

func (b *Balloon) Add(event []byte) (*Commitment, error) {
	digest := b.hasher(event)
	b.version++
	index := make([]byte, 8)
	binary.LittleEndian.PutUint64(index, uint64(b.version))

	b.store.Add(digest, index)

	historyAddOp, historyAddResult := history.NewAdd(digest, index)
	hyperAddOp, hyperAddResult := hyper.NewAdd(digest, index)

	b.history <- historyAddOp
	b.hyper <- hyperAddOp

	historyDigest := <-historyAddResult
	hyperDigest := <-hyperAddResult

	return &Commitment{historyDigest, hyperDigest, b.version}, nil
}
