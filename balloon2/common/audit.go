package common

import "github.com/aalda/trees/log"

type AuditPath map[string]Digest

func (p AuditPath) Get(pos Position) (Digest, bool) {
	digest, ok := p[pos.StringId()]
	return digest, ok
}

type Verifiable interface {
	Verify(expectedDigest Digest, key, value []byte) bool
	AuditPath() AuditPath
}

type FakeVerifiable struct {
	result bool
}

func NewFakeVerifiable(result bool) *FakeVerifiable {
	return &FakeVerifiable{result}
}

func (f FakeVerifiable) Verify(commitment Digest, key, value []byte) bool {
	return f.result
}

func (f FakeVerifiable) AuditPath() AuditPath {
	return make(AuditPath)
}

type AuditPathVisitor struct {
	decorated *ComputeHashVisitor
	auditPath AuditPath
}

func NewAuditPathVisitor(decorated *ComputeHashVisitor) *AuditPathVisitor {
	return &AuditPathVisitor{decorated, make(AuditPath)}
}

func (v AuditPathVisitor) Result() AuditPath {
	return v.auditPath
}

func (v *AuditPathVisitor) VisitRoot(pos Position, leftResult, rightResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitRoot(pos, leftResult, rightResult)
}

func (v *AuditPathVisitor) VisitNode(pos Position, leftResult, rightResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitNode(pos, leftResult, rightResult)
}

func (v *AuditPathVisitor) VisitPartialNode(pos Position, leftResult interface{}) interface{} {
	// by-pass
	return v.decorated.VisitPartialNode(pos, leftResult)
}

func (v *AuditPathVisitor) VisitLeaf(pos Position, eventDigest []byte) interface{} {
	// ignore. target leafs not included in path
	return v.decorated.VisitLeaf(pos, eventDigest)
}

func (v *AuditPathVisitor) VisitCached(pos Position, cachedDigest Digest) interface{} {
	// by-pass
	return v.decorated.VisitCached(pos, cachedDigest)
}

func (v *AuditPathVisitor) VisitCollectable(pos Position, result interface{}) interface{} {
	digest := v.decorated.VisitCollectable(pos, result)
	log.Debugf("Adding collectable to path in position: %v", pos)
	v.auditPath[pos.StringId()] = digest.(Digest)
	return digest
}
