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
	history *history.Tree
	hyper   *hyper.Tree
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
		history,
		hyper,
		hasher,
		0,
	}

	return &b

}

func (b *Balloon) Add(event []byte) (*Commitment, error) {
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

func (b *Balloon) Close() error {
	b.history.Close()
	b.hyper.Close()

	return nil
}
