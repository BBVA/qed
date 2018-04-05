package balloon

import (
	"encoding/binary"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/merkle"
	"verifiabledata/balloon/storage"
)

type Balloon struct {
	history *merkle.TreeChannel
	hyper   *merkle.TreeChannel
	store   storage.Store
	hasher  hashing.Hasher
	version uint
}

type Commitment struct {
	HistoryDigest []byte
	IndexDigest   []byte
	version       uint
}

func NewBalloon(store storage.Store, hasher hashing.Hasher) *Balloon {

	htChannel := merkle.NewTreeChannel()
	hyperChannel := merkle.NewTreeChannel()

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
	go b.history.Run(b.history)
	go b.hyper.Run(b.hyper)
}

func (b *Balloon) Add(event []byte) (*Commitment, error) {
	digest := b.hasher(event)
	b.version++
	index := asBytes(b.version)

	b.store.Add(digest, index)

	kvPair := &KVPair{digest, index}

	b.history.send <- kvPair
	b.hyper.send <- kvPair

	historyDigest := <-b.history.receive
	hyperDigest := <-b.hyper.receive

	return &Commitment{historyDigest, hyperDigest, b.version}, nil
}

func asBytes(value uint) []byte {
	valueBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueBytes, uint32(value))
	return valueBytes
}
