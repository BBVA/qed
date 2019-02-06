package pruning5

import (
	"fmt"

	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
)

type Operation interface {
	Accept(visitor OpVisitor) hashing.Digest
	String() string
	Position() *navigation.Position
}

type LeafHashOp struct {
	pos   *navigation.Position
	Value []byte
}

type InnerHashOp struct {
	pos         *navigation.Position
	Left, Right Operation
}

type PartialInnerHashOp struct {
	pos  *navigation.Position
	Left Operation
}

type GetCacheOp struct {
	pos *navigation.Position
}

type PutCacheOp struct {
	Operation
}

type MutateOp struct {
	Operation
}

type CollectOp struct {
	Operation
}

func NewLeafHashOp(pos *navigation.Position, value []byte) *LeafHashOp {
	return &LeafHashOp{
		pos:   pos,
		Value: value,
	}
}

func (o LeafHashOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitLeafHashOp(o)
}

func (o LeafHashOp) Position() *navigation.Position {
	return o.pos
}

func (o LeafHashOp) String() string {
	return fmt.Sprintf("LeafHashOp(%v)[ %x ]", o.pos, o.Value)
}

func NewInnerHashOp(pos *navigation.Position, left, right Operation) *InnerHashOp {
	return &InnerHashOp{
		pos:   pos,
		Left:  left,
		Right: right,
	}
}

func (o InnerHashOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitInnerHashOp(o)
}

func (o InnerHashOp) Position() *navigation.Position {
	return o.pos
}

func (o InnerHashOp) String() string {
	return fmt.Sprintf("InnerHashOp(%v)[ %v | %v ]", o.pos, o.Left, o.Right)
}

func NewPartialInnerHashOp(pos *navigation.Position, left Operation) *PartialInnerHashOp {
	return &PartialInnerHashOp{
		pos:  pos,
		Left: left,
	}
}

func (o PartialInnerHashOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitPartialInnerHashOp(o)
}

func (o PartialInnerHashOp) Position() *navigation.Position {
	return o.pos
}

func (o PartialInnerHashOp) String() string {
	return fmt.Sprintf("PartialInnerHashOp(%v)[ %v ]", o.pos, o.Left)
}

func NewGetCacheOp(pos *navigation.Position) *GetCacheOp {
	return &GetCacheOp{
		pos: pos,
	}
}

func (o GetCacheOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitGetCacheOp(o)
}

func (o GetCacheOp) Position() *navigation.Position {
	return o.pos
}

func (o GetCacheOp) String() string {
	return fmt.Sprintf("GetCacheOp(%v)", o.pos)
}

func NewPutCacheOp(op Operation) *PutCacheOp {
	return &PutCacheOp{
		Operation: op,
	}
}

func (o PutCacheOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitPutCacheOp(o)
}

func (o PutCacheOp) Position() *navigation.Position {
	return o.Operation.Position()
}

func (o PutCacheOp) String() string {
	return fmt.Sprintf("PutCacheOp[ %v ]", o.Operation)
}

func NewMutateOp(op Operation) *MutateOp {
	return &MutateOp{
		Operation: op,
	}
}

func (o MutateOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitMutateOp(o)
}

func (o MutateOp) Position() *navigation.Position {
	return o.Operation.Position()
}

func (o MutateOp) String() string {
	return fmt.Sprintf("MutateOp[ %v ]", o.Operation)
}

func NewCollectOp(op Operation) *CollectOp {
	return &CollectOp{
		Operation: op,
	}
}

func (o CollectOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitCollectOp(o)
}

func (o CollectOp) Position() *navigation.Position {
	return o.Operation.Position()
}

func (o CollectOp) String() string {
	return fmt.Sprintf("CollectOp[ %v ]", o.Operation)
}
