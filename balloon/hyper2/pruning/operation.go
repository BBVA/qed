package pruning

import (
	"fmt"

	"github.com/bbva/qed/balloon/hyper2/navigation"
	"github.com/bbva/qed/hashing"
)

type Operation interface {
	Accept(visitor OpVisitor) hashing.Digest
	String() string
	Position() *navigation.Position
}

type OpVisitor interface { // THROW ERRORS?
	VisitShortcutLeafOp(op ShortcutLeafOp) hashing.Digest
	VisitLeafOp(op LeafOp) hashing.Digest
	VisitInnerHashOp(op InnerHashOp) hashing.Digest
	VisitGetDefaultOp(op GetDefaultOp) hashing.Digest
	VisitUseProvidedOp(op UseProvidedOp) hashing.Digest
	VisitPutBatchOp(op PutBatchOp) hashing.Digest
	VisitMutateBatchOp(op MutateBatchOp) hashing.Digest
	VisitCollectOp(op CollectOp) hashing.Digest
}

type ShortcutLeafOp struct {
	pos        *navigation.Position
	Batch      *BatchNode
	Idx        int8
	Key, Value []byte
}

type LeafOp struct {
	pos   *navigation.Position
	Batch *BatchNode
	Idx   int8
	Operation
}

type InnerHashOp struct {
	pos         *navigation.Position
	Batch       *BatchNode
	Idx         int8
	Left, Right Operation
}

type GetDefaultOp struct {
	pos *navigation.Position
}

type UseProvidedOp struct {
	pos   *navigation.Position
	Batch *BatchNode
	Idx   int8
}

type PutBatchOp struct {
	Operation
	Batch *BatchNode
}

type MutateBatchOp struct {
	Operation
	Batch *BatchNode
}

type CollectOp struct {
	Operation
}

func NewShortcutLeafOp(pos *navigation.Position, batch *BatchNode, iBatch int8, key, value []byte) *ShortcutLeafOp {
	return &ShortcutLeafOp{
		pos:   pos,
		Batch: batch,
		Idx:   iBatch,
		Key:   key,
		Value: value,
	}
}

func (o ShortcutLeafOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitShortcutLeafOp(o)
}

func (o ShortcutLeafOp) Position() *navigation.Position {
	return o.pos
}

func (o ShortcutLeafOp) String() string {
	return fmt.Sprintf("ShortcutLeafOp(%v)(%d)[ %x - %x ]", o.pos, o.Idx, o.Key, o.Value)
}

func NewLeafOp(pos *navigation.Position, batch *BatchNode, iBatch int8, op Operation) *LeafOp {
	return &LeafOp{
		pos:       pos,
		Batch:     batch,
		Idx:       iBatch,
		Operation: op,
	}
}

func (o LeafOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitLeafOp(o)
}

func (o LeafOp) Position() *navigation.Position {
	return o.pos
}

func (o LeafOp) String() string {
	return fmt.Sprintf("LeafOp(%v)(%d)[ %v ]", o.pos, o.Idx, o.Operation)
}

func NewInnerHashOp(pos *navigation.Position, batch *BatchNode, iBatch int8, left, right Operation) *InnerHashOp {
	return &InnerHashOp{
		pos:   pos,
		Batch: batch,
		Idx:   iBatch,
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
	return fmt.Sprintf("InnerHashOp(%v)(%d)[ %v | %v ]", o.pos, o.Idx, o.Left, o.Right)
}

func NewGetDefaultOp(pos *navigation.Position) *GetDefaultOp {
	return &GetDefaultOp{
		pos: pos,
	}
}

func (o GetDefaultOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitGetDefaultOp(o)
}

func (o GetDefaultOp) Position() *navigation.Position {
	return o.pos
}

func (o GetDefaultOp) String() string {
	return fmt.Sprintf("GetDefaultOp(%v)", o.pos)
}

func NewUseProvidedOp(pos *navigation.Position, batch *BatchNode, iBatch int8) *UseProvidedOp {
	return &UseProvidedOp{
		pos:   pos,
		Batch: batch,
		Idx:   iBatch,
	}
}

func (o UseProvidedOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitUseProvidedOp(o)
}

func (o UseProvidedOp) Position() *navigation.Position {
	return o.pos
}

func (o UseProvidedOp) String() string {
	return fmt.Sprintf("UseProvidedOp(%v)(%d)", o.pos, o.Idx)
}

func NewPutBatchOp(op Operation, batch *BatchNode) *PutBatchOp {
	return &PutBatchOp{
		Operation: op,
		Batch:     batch,
	}
}

func (o PutBatchOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitPutBatchOp(o)
}

func (o PutBatchOp) Position() *navigation.Position {
	return o.Operation.Position()
}

func (o PutBatchOp) String() string {
	return fmt.Sprintf("PutBatchOp( %v )[ %v ]", o.Operation, o.Batch)
}

func NewMutateBatchOp(op Operation, batch *BatchNode) *MutateBatchOp {
	return &MutateBatchOp{
		Operation: op,
		Batch:     batch,
	}
}

func (o MutateBatchOp) Accept(visitor OpVisitor) hashing.Digest {
	return visitor.VisitMutateBatchOp(o)
}

func (o MutateBatchOp) Position() *navigation.Position {
	return o.Operation.Position()
}

func (o MutateBatchOp) String() string {
	return fmt.Sprintf("MutateBatchOp( %v )[ %v ]", o.Operation, o.Batch)
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
	return fmt.Sprintf("CollectOp( %v )", o.Operation)
}
