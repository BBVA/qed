package common

import (
	"fmt"
)

type PostOrderVisitor interface {
	VisitRoot(pos Position, leftResult, rightResult interface{}) interface{}
	VisitNode(pos Position, leftResult, rightResult interface{}) interface{}
	VisitPartialNode(pos Position, leftResult interface{}) interface{}
	VisitLeaf(pos Position, value []byte) interface{}
	VisitCached(pos Position, cachedDigest Digest) interface{}
	VisitCollectable(pos Position, result interface{}) interface{}
}

type PreOrderVisitor interface {
	VisitRoot(pos Position)
	VisitNode(pos Position)
	VisitPartialNode(pos Position)
	VisitLeaf(pos Position, value []byte)
	VisitCached(pos Position, cachedDigest Digest)
	VisitCollectable(pos Position)
}

type Visitable interface {
	PostOrder(visitor PostOrderVisitor) interface{}
	PreOrder(visitor PreOrderVisitor)
	String() string
	Position() Position
}

type Root struct {
	pos         Position
	left, right Visitable
}

type Node struct {
	pos         Position
	left, right Visitable
}

type PartialNode struct {
	pos  Position
	left Visitable
}

type Leaf struct {
	pos   Position
	value []byte
}

type Cached struct {
	pos    Position
	digest Digest
}

type Collectable struct {
	underlying Visitable
}

func NewRoot(pos Position, left, right Visitable) *Root {
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

func (r Root) Position() Position {
	return r.pos
}

func (r Root) String() string {
	return fmt.Sprintf("Root(%v)[ %v | %v ]", r.pos, r.left, r.right)
}

func NewNode(pos Position, left, right Visitable) *Node {
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

func (n Node) Position() Position {
	return n.pos
}

func (n Node) String() string {
	return fmt.Sprintf("Node(%v)[ %v | %v ]", n.pos, n.left, n.right)
}

func NewPartialNode(pos Position, left Visitable) *PartialNode {
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

func (p PartialNode) Position() Position {
	return p.pos
}

func (p PartialNode) String() string {
	return fmt.Sprintf("PartialNode(%v)[ %v ]", p.pos, p.left)
}

func NewLeaf(pos Position, value []byte) *Leaf {
	return &Leaf{pos, value}
}

func (l Leaf) PostOrder(visitor PostOrderVisitor) interface{} {
	return visitor.VisitLeaf(l.pos, l.value)
}

func (l Leaf) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitLeaf(l.pos, l.value)
}

func (l Leaf) Position() Position {
	return l.pos
}

func (l Leaf) String() string {
	return fmt.Sprintf("Leaf(%v)[ %x ]", l.pos, l.value)
}

func NewCached(pos Position, digest Digest) *Cached {
	return &Cached{pos, digest}
}

func (c Cached) PostOrder(visitor PostOrderVisitor) interface{} {
	return visitor.VisitCached(c.pos, c.digest)
}

func (c Cached) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitCached(c.pos, c.digest)
}

func (c Cached) Position() Position {
	return c.pos
}

func (c Cached) String() string {
	return fmt.Sprintf("Cached(%v)[ %x ]", c.pos, c.digest)
}

func NewCollectable(underlying Visitable) *Collectable {
	return &Collectable{underlying}
}

func (c Collectable) PostOrder(visitor PostOrderVisitor) interface{} {
	result := c.underlying.PostOrder(visitor)
	return visitor.VisitCollectable(c.Position(), result)
}

func (c Collectable) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitCollectable(c.Position())
	c.underlying.PreOrder(visitor)
}

func (c Collectable) Position() Position {
	return c.underlying.Position()
}

func (c Collectable) String() string {
	return fmt.Sprintf("Collectable[ %v ]", c.underlying)
}
