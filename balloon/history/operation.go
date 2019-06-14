/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package history

import (
	"fmt"

	"github.com/bbva/qed/crypto/hashing"
)

type operation interface {
	Accept(visitor opVisitor) hashing.Digest
	String() string
	Position() *position
}

type opVisitor interface { // THROW ERRORS?
	VisitLeafHashOp(op leafHashOp) hashing.Digest
	VisitInnerHashOp(op innerHashOp) hashing.Digest
	VisitPartialInnerHashOp(op partialInnerHashOp) hashing.Digest
	VisitGetCacheOp(op getCacheOp) hashing.Digest
	VisitPutCacheOp(op putCacheOp) hashing.Digest
	VisitMutateOp(op mutateOp) hashing.Digest
	VisitCollectOp(op collectOp) hashing.Digest
}

type leafHashOp struct {
	pos   *position
	Value []byte
}

type innerHashOp struct {
	pos         *position
	Left, Right operation
}

type partialInnerHashOp struct {
	pos  *position
	Left operation
}

type getCacheOp struct {
	pos *position
}

type putCacheOp struct {
	operation
}

type mutateOp struct {
	operation
}

type collectOp struct {
	operation
}

func newLeafHashOp(pos *position, value []byte) *leafHashOp {
	return &leafHashOp{
		pos:   pos,
		Value: value,
	}
}

func (o leafHashOp) Accept(visitor opVisitor) hashing.Digest {
	return visitor.VisitLeafHashOp(o)
}

func (o leafHashOp) Position() *position {
	return o.pos
}

func (o leafHashOp) String() string {
	return fmt.Sprintf("leafHashOp(%v)[ %x ]", o.pos, o.Value)
}

func newInnerHashOp(pos *position, left, right operation) *innerHashOp {
	return &innerHashOp{
		pos:   pos,
		Left:  left,
		Right: right,
	}
}

func (o innerHashOp) Accept(visitor opVisitor) hashing.Digest {
	return visitor.VisitInnerHashOp(o)
}

func (o innerHashOp) Position() *position {
	return o.pos
}

func (o innerHashOp) String() string {
	return fmt.Sprintf("innerHashOp(%v)[ %v | %v ]", o.pos, o.Left, o.Right)
}

func newPartialInnerHashOp(pos *position, left operation) *partialInnerHashOp {
	return &partialInnerHashOp{
		pos:  pos,
		Left: left,
	}
}

func (o partialInnerHashOp) Accept(visitor opVisitor) hashing.Digest {
	return visitor.VisitPartialInnerHashOp(o)
}

func (o partialInnerHashOp) Position() *position {
	return o.pos
}

func (o partialInnerHashOp) String() string {
	return fmt.Sprintf("partialInnerHashOp(%v)[ %v ]", o.pos, o.Left)
}

func newGetCacheOp(pos *position) *getCacheOp {
	return &getCacheOp{
		pos: pos,
	}
}

func (o getCacheOp) Accept(visitor opVisitor) hashing.Digest {
	return visitor.VisitGetCacheOp(o)
}

func (o getCacheOp) Position() *position {
	return o.pos
}

func (o getCacheOp) String() string {
	return fmt.Sprintf("getCacheOp(%v)", o.pos)
}

func newPutCacheOp(op operation) *putCacheOp {
	return &putCacheOp{
		operation: op,
	}
}

func (o putCacheOp) Accept(visitor opVisitor) hashing.Digest {
	return visitor.VisitPutCacheOp(o)
}

func (o putCacheOp) Position() *position {
	return o.operation.Position()
}

func (o putCacheOp) String() string {
	return fmt.Sprintf("putCacheOp( %v )", o.operation)
}

func newMutateOp(op operation) *mutateOp {
	return &mutateOp{
		operation: op,
	}
}

func (o mutateOp) Accept(visitor opVisitor) hashing.Digest {
	return visitor.VisitMutateOp(o)
}

func (o mutateOp) Position() *position {
	return o.operation.Position()
}

func (o mutateOp) String() string {
	return fmt.Sprintf("mutateOp( %v )", o.operation)
}

func newCollectOp(op operation) *collectOp {
	return &collectOp{
		operation: op,
	}
}

func (o collectOp) Accept(visitor opVisitor) hashing.Digest {
	return visitor.VisitCollectOp(o)
}

func (o collectOp) Position() *position {
	return o.operation.Position()
}

func (o collectOp) String() string {
	return fmt.Sprintf("collectOp( %v )", o.operation)
}
