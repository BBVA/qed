package common

import (
	"encoding/binary"
	"fmt"

	"github.com/bbva/qed/util"
)

type FakePosition struct {
	index  []byte
	height uint16
}

func (p FakePosition) Index() []byte {
	return p.index
}

func (p FakePosition) Height() uint16 {
	return p.height
}

func (p FakePosition) Bytes() []byte {
	b := make([]byte, 34) // Size of the index plus 2 bytes for the height
	copy(b, p.index)
	copy(b[len(p.index):], util.Uint16AsBytes(p.height))
	return b
}

func (p FakePosition) String() string {
	return fmt.Sprintf("Pos(%x, %d)", p.index, p.height)
}

func (p FakePosition) StringId() string {
	return fmt.Sprintf("%x|%d", p.index, p.height)
}

func (p FakePosition) IndexAsUint64() uint64 {
	return binary.BigEndian.Uint64(p.index)
}
