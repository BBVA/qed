package history

import (
	"fmt"
	"verifiabledata/util"
)

// Position holds the index and the layer of a node in a tree
type Position struct {
	Index uint64
	Layer uint64
}

func (p *Position) GetBytes() []byte {
	return append(p.IndexAsBytes(), p.LayerAsBytes()...)
}

func (p *Position) String() string {
	return fmt.Sprintf("(i %d, l %d)", p.Index, p.Layer)
}

func (p *Position) IndexAsBytes() []byte {
	return util.UInt64AsBytes(p.Index)
}

func (p *Position) LayerAsBytes() []byte {
	return util.UInt64AsBytes(p.Layer)
}

// Utility to allocate a new Position
func newPosition(index, layer uint64) *Position {
	return &Position{index, layer}
}
