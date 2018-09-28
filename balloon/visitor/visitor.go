/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

package visitor

import (
	"fmt"

	"github.com/bbva/qed/balloon/navigator"
	"github.com/bbva/qed/hashing"
)

type PostOrderVisitor interface {
	VisitRoot(pos navigator.Position, leftResult, rightResult interface{}) interface{}
	VisitNode(pos navigator.Position, leftResult, rightResult interface{}) interface{}
	VisitPartialNode(pos navigator.Position, leftResult interface{}) interface{}
	VisitLeaf(pos navigator.Position, value []byte) interface{}
	VisitCached(pos navigator.Position, cachedDigest hashing.Digest) interface{}
	VisitCollectable(pos navigator.Position, result interface{}) interface{}
	VisitCacheable(pos navigator.Position, result interface{}) interface{}
}

type PreOrderVisitor interface {
	VisitRoot(pos navigator.Position)
	VisitNode(pos navigator.Position)
	VisitPartialNode(pos navigator.Position)
	VisitLeaf(pos navigator.Position, value []byte)
	VisitCached(pos navigator.Position, cachedDigest hashing.Digest)
	VisitCollectable(pos navigator.Position)
	VisitCacheable(pos navigator.Position)
}

type Visitable interface {
	PostOrder(visitor PostOrderVisitor) interface{}
	PreOrder(visitor PreOrderVisitor)
	String() string
	Position() navigator.Position
}

type Root struct {
	pos         navigator.Position
	left, right Visitable
}

type Node struct {
	pos         navigator.Position
	left, right Visitable
}

type PartialNode struct {
	pos  navigator.Position
	left Visitable
}

type Leaf struct {
	pos   navigator.Position
	value []byte
}

type Cached struct {
	pos    navigator.Position
	digest hashing.Digest
}

type Cacheable struct {
	Visitable
}

type Collectable struct {
	Visitable
}

func NewRoot(pos navigator.Position, left, right Visitable) *Root {
	return &Root{pos, left, right}
}

func (r Root) PostOrder(visitor PostOrderVisitor) interface{} {
	leftResult := r.left.PostOrder(visitor)
	rightResult := r.right.PostOrder(visitor)
	return visitor.VisitRoot(r.pos, leftResult, rightResult)
}

func (r Root) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitRoot(r.pos)
	r.left.PreOrder(visitor)
	r.right.PreOrder(visitor)
}

func (r Root) Position() navigator.Position {
	return r.pos
}

func (r Root) String() string {
	return fmt.Sprintf("Root(%v)[ %v | %v ]", r.pos, r.left, r.right)
}

func NewNode(pos navigator.Position, left, right Visitable) *Node {
	return &Node{pos, left, right}
}

func (n Node) PostOrder(visitor PostOrderVisitor) interface{} {
	leftResult := n.left.PostOrder(visitor)
	rightResult := n.right.PostOrder(visitor)
	return visitor.VisitNode(n.pos, leftResult, rightResult)
}

func (n Node) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitNode(n.pos)
	n.left.PreOrder(visitor)
	n.right.PreOrder(visitor)
}

func (n Node) Position() navigator.Position {
	return n.pos
}

func (n Node) String() string {
	return fmt.Sprintf("Node(%v)[ %v | %v ]", n.pos, n.left, n.right)
}

func NewPartialNode(pos navigator.Position, left Visitable) *PartialNode {
	return &PartialNode{pos, left}
}

func (p PartialNode) PostOrder(visitor PostOrderVisitor) interface{} {
	leftResult := p.left.PostOrder(visitor)
	return visitor.VisitPartialNode(p.pos, leftResult)
}

func (p PartialNode) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitPartialNode(p.pos)
	p.left.PreOrder(visitor)
}

func (p PartialNode) Position() navigator.Position {
	return p.pos
}

func (p PartialNode) String() string {
	return fmt.Sprintf("PartialNode(%v)[ %v ]", p.pos, p.left)
}

func NewLeaf(pos navigator.Position, value []byte) *Leaf {
	return &Leaf{pos, value}
}

func (l Leaf) PostOrder(visitor PostOrderVisitor) interface{} {
	return visitor.VisitLeaf(l.pos, l.value)
}

func (l Leaf) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitLeaf(l.pos, l.value)
}

func (l Leaf) Position() navigator.Position {
	return l.pos
}

func (l Leaf) String() string {
	return fmt.Sprintf("Leaf(%v)[ %x ]", l.pos, l.value)
}

func NewCached(pos navigator.Position, digest hashing.Digest) *Cached {
	return &Cached{pos, digest}
}

func (c Cached) PostOrder(visitor PostOrderVisitor) interface{} {
	return visitor.VisitCached(c.pos, c.digest)
}

func (c Cached) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitCached(c.pos, c.digest)
}

func (c Cached) Position() navigator.Position {
	return c.pos
}

func (c Cached) String() string {
	return fmt.Sprintf("Cached(%v)[ %x ]", c.pos, c.digest)
}

func NewCacheable(underlying Visitable) *Cacheable {
	return &Cacheable{Visitable: underlying}
}

func (c Cacheable) PostOrder(visitor PostOrderVisitor) interface{} {
	result := c.Visitable.PostOrder(visitor)
	return visitor.VisitCacheable(c.Position(), result)
}

func (c Cacheable) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitCacheable(c.Position())
	c.Visitable.PreOrder(visitor)
}

func (c Cacheable) String() string {
	return fmt.Sprintf("Cacheable[ %v ]", c.Visitable)
}

func NewCollectable(underlying Visitable) *Collectable {
	return &Collectable{Visitable: underlying}
}

func (c Collectable) PostOrder(visitor PostOrderVisitor) interface{} {
	result := c.Visitable.PostOrder(visitor)
	return visitor.VisitCollectable(c.Position(), result)
}

func (c Collectable) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitCollectable(c.Position())
	c.Visitable.PreOrder(visitor)
}

func (c Collectable) String() string {
	return fmt.Sprintf("Collectable[ %v ]", c.Visitable)
}
