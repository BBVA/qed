package balloon

import (
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/merkle"
	"verifiabledata/balloon/storage"
	"verifiabledata/util"
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

func NewBalloon(store storage.Store, hasher hashing.Hasher) *Balloon {

	htChannel := make(chan interface{})
	hyperChannel := make(chan interface{})

	history := history.NewTree()
	hyper := hyper.NewTree()

	return &Balloon{
		htChannel,
		hyperChannel,
		store,
		hasher,
		0,
	}
}

func (b *Balloon) Start() error {
	b.history.Run(b.history)
	b.hyper.Run(b.hyper)
}


func (b *Balloon) Add(event []byte) (*Commitment, error) {
	digest := b.hasher(event)
	b.version++
	index := util.asBytes(b.version)

	b.store.Add(digest, index)

	historyAddOp, historyAddResult  := history.NewAdd(digest, index)
	hyperAddOp, hyperAddResult := hyper.NewAdd(digest, index)
	
	b.history <- historyAddOp
	b.hyper <- hyperAddOp


	historyDigest := <-historyAddResult
	hyperDigest := <- hyperAddResult

	return &Commitment{historyDigest, hyperDigest, b.version}, nil
}
