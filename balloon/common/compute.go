package common

import "github.com/bbva/qed/log"

type ComputeHashVisitor struct {
	hasher Hasher
}

func NewComputeHashVisitor(hasher Hasher) *ComputeHashVisitor {
	return &ComputeHashVisitor{hasher}
}

func (v *ComputeHashVisitor) VisitRoot(pos Position, leftResult, rightResult interface{}) interface{} {
	log.Debugf("Computing root hash in position: %v", pos)
	return v.interiorHash(pos.Bytes(), leftResult.(Digest), rightResult.(Digest))
}

func (v *ComputeHashVisitor) VisitNode(pos Position, leftResult, rightResult interface{}) interface{} {
	log.Debugf("Computing node hash in position: %v", pos)
	return v.interiorHash(pos.Bytes(), leftResult.(Digest), rightResult.(Digest))
}

func (v *ComputeHashVisitor) VisitPartialNode(pos Position, leftResult interface{}) interface{} {
	log.Debugf("Computing partial node hash in position: %v", pos)
	return v.leafHash(pos.Bytes(), leftResult.(Digest))
}

func (v *ComputeHashVisitor) VisitLeaf(pos Position, value []byte) interface{} {
	log.Debugf("Computing leaf hash in position: %v", pos)
	return v.leafHash(pos.Bytes(), value)
}

func (v *ComputeHashVisitor) VisitCached(pos Position, cachedDigest Digest) interface{} {
	log.Debugf("Getting cached hash in position: %v", pos)
	return cachedDigest
}

func (v *ComputeHashVisitor) VisitCollectable(pos Position, result interface{}) interface{} {
	log.Debugf("Getting collectable value in position: %v", pos)
	return result
}

func (v *ComputeHashVisitor) leafHash(id, leaf []byte) Digest {
	return v.hasher.Salted(id, leaf)
}

func (v *ComputeHashVisitor) interiorHash(id, left, right []byte) Digest {
	return v.hasher.Salted(id, left, right)
}
