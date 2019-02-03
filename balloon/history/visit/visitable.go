package visit

import (
	"fmt"

	"github.com/bbva/qed/balloon/history/navigation"
	"github.com/bbva/qed/hashing"
)

type Visitable interface {
	PostOrder(visitor PostOrderVisitor) hashing.Digest
	PreOrder(visitor PreOrderVisitor)
	String() string
	Position() *navigation.Position
}

type Node struct {
	pos         *navigation.Position
	left, right Visitable
}

type PartialNode struct {
	pos  *navigation.Position
	left Visitable
}

type Leaf struct {
	pos   *navigation.Position
	value []byte
}

type Cached struct {
	pos    *navigation.Position
	digest hashing.Digest
}

type Cacheable struct {
	Visitable
}

type Mutable struct {
	Visitable
}

type Collectable struct {
	Visitable
}

func NewNode(pos *navigation.Position, left, right Visitable) *Node {
	return &Node{
		pos:   pos,
		left:  left,
		right: right,
	}
}

func (n Node) PostOrder(visitor PostOrderVisitor) hashing.Digest {
	leftResult := n.left.PostOrder(visitor)
	rightResult := n.right.PostOrder(visitor)
	return visitor.VisitNode(n.pos, leftResult, rightResult)
}

func (n Node) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitNode(n.pos)
	n.left.PreOrder(visitor)
	n.right.PreOrder(visitor)
}

func (n Node) Position() *navigation.Position {
	return n.pos
}

func (n Node) String() string {
	return fmt.Sprintf("Node(%v)[ %v | %v ]", n.pos, n.left, n.right)
}

func NewPartialNode(pos *navigation.Position, left Visitable) *PartialNode {
	return &PartialNode{
		pos:  pos,
		left: left,
	}
}

func (p PartialNode) PostOrder(visitor PostOrderVisitor) hashing.Digest {
	leftResult := p.left.PostOrder(visitor)
	return visitor.VisitPartialNode(p.pos, leftResult)
}

func (p PartialNode) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitPartialNode(p.pos)
	p.left.PreOrder(visitor)
}

func (p PartialNode) Position() *navigation.Position {
	return p.pos
}

func (p PartialNode) String() string {
	return fmt.Sprintf("PartialNode(%v)[ %v ]", p.pos, p.left)
}

func NewLeaf(pos *navigation.Position, value []byte) *Leaf {
	return &Leaf{
		pos:   pos,
		value: value,
	}
}

func (l Leaf) PostOrder(visitor PostOrderVisitor) hashing.Digest {
	return visitor.VisitLeaf(l.pos, l.value)
}

func (l Leaf) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitLeaf(l.pos, l.value)
}

func (l Leaf) Position() *navigation.Position {
	return l.pos
}

func (l Leaf) String() string {
	return fmt.Sprintf("Leaf(%v)[ %x ]", l.pos, l.value)
}

func NewCached(pos *navigation.Position, digest hashing.Digest) *Cached {
	return &Cached{
		pos:    pos,
		digest: digest,
	}
}

func (c Cached) PostOrder(visitor PostOrderVisitor) hashing.Digest {
	return visitor.VisitCached(c.pos, c.digest)
}

func (c Cached) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitCached(c.pos, c.digest)
}

func (c Cached) Position() *navigation.Position {
	return c.pos
}

func (c Cached) String() string {
	return fmt.Sprintf("Cached(%v)[ %x ]", c.pos, c.digest)
}

func NewCacheable(underlying Visitable) *Cacheable {
	return &Cacheable{Visitable: underlying}
}

func (c Cacheable) PostOrder(visitor PostOrderVisitor) hashing.Digest {
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

func NewMutable(underlying Visitable) *Mutable {
	return &Mutable{underlying}
}

func (c Mutable) PostOrder(visitor PostOrderVisitor) hashing.Digest {
	result := c.Visitable.PostOrder(visitor)
	return visitor.VisitMutable(c.Position(), result)
}

func (c Mutable) PreOrder(visitor PreOrderVisitor) {
	visitor.VisitMutable(c.Position())
	c.Visitable.PreOrder(visitor)
}

func (c Mutable) String() string {
	return fmt.Sprintf("Mutable[ %v ]", c.Visitable)
}

func NewCollectable(underlying Visitable) *Collectable {
	return &Collectable{underlying}
}

func (c Collectable) PostOrder(visitor PostOrderVisitor) hashing.Digest {
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
