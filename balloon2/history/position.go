package history

import (
	"fmt"

	"github.com/bbva/qed/util"
)

type HistoryPosition struct {
	index  uint64
	height uint16
}

func NewPosition(index uint64, height uint16) *HistoryPosition {
	return &HistoryPosition{
		index:  index,
		height: height,
	}
}

func (p HistoryPosition) Index() []byte {
	return util.Uint64AsBytes(p.index)
}

func (p HistoryPosition) Height() uint16 {
	return p.height
}

func (p HistoryPosition) IndexAsUint64() uint64 {
	return p.index
}

func (p HistoryPosition) Bytes() []byte {
	b := make([]byte, 10) // Size of the index plus 2 bytes for the height
	indexAsBytes := p.Index()
	copy(b, indexAsBytes)
	copy(b[len(indexAsBytes):], util.Uint16AsBytes(p.height))
	return b
}

func (p HistoryPosition) String() string {
	return fmt.Sprintf("Pos(%d, %d)", p.IndexAsUint64(), p.height)
}

func (p HistoryPosition) StringId() string {
	return fmt.Sprintf("%d|%d", p.IndexAsUint64(), p.height)
}
