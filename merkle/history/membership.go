// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

/*
	Package history implements a history tree structure as described in the paper
	    Balloon: A Forward-Secure Append-Only Persistent Authenticated Data Structure
	    https://eprint.iacr.org/2015/007
*/
package history

import (
	"errors"

	"verifiabledata/util"
)

type MembershipProof struct {
	Index   uint64
	Layer   uint64
	Version uint64
	Hash    []byte
	Root    *Node
	Nodes   []Node
}

// MembershipProof generates a membership proof.
func (t *Tree) MembershipProof(index, version uint64, hash []byte) (proof MembershipProof, err error) {
	if index < 0 || index >= t.size || index > version {
		return proof, errors.New("invalid index, has to be: 0 <= index <= version < size")
	}

	proof.Index = index
	proof.Version = version
	proof.Hash = hash

	proof.Root, err = t.getNode(0, t.getDepth(version+1), version)
	if err != nil {
		return
	}

	// we know that the biggest possible proof is one node per layer
	proof.Nodes = make([]Node, 0, t.getDepth(t.size))
	err = t.membershipProof(index, 0, t.getDepth(t.size), version, &proof)
	return
}

// the game is: walk the tree from the root to the target leaf
func (t *Tree) membershipProof(target, index, layer, version uint64, proof *MembershipProof) (err error) {
	if layer == 0 {
		return
	}

	// the number of events to the left of the node
	n := index + util.Pow(2, layer-1)

	if target < n {
		// go left, but should we save right first? We need to save right if
		// there are any leaf nodes fixed by the right node (otherwise we
		// know it's hash is nil), dictated by the version of the tree we
		// are generating
		if version >= n {
			node, err := t.getNode(n, layer-1, version)
			if err != nil {
				return err
			}

			proof.Nodes = append(proof.Nodes, *node)
		}

		return t.membershipProof(target, index, layer-1, version, proof)
	}

	// go right, once we have saved the left node
	node, err := t.getNode(index, layer-1, version)
	if err != nil {
		return err
	}

	proof.Nodes = append(proof.Nodes, *node)

	return t.membershipProof(target, n, layer-1, version, proof)
}

// Verify verifies a membership proof
func (p *MembershipProof) Verify() (correct bool) {
	if p.Root == nil || p.Hash == nil || p.Index < 0 || p.Version < 0 {
		return false
	}

	proofTree := NewInmemoryTree()
	for _, node := range p.Nodes {
		err := proofTree.frozen.Add(&node)
		if err != nil {
			panic(err)
			// return false
		}
	}

	err := proofTree.events.Add(&Node{&Position{p.Index, p.Layer}, p.Hash})
	if err != nil {
		panic(err)
		// return false
	}

	commitment, err := proofTree.getNode(0, proofTree.getDepth(p.Version+1), p.Version)
	if err != nil {
		return false
	}

	return util.Equal(commitment.Digest, p.Root.Digest)
}
