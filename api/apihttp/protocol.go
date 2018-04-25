package apihttp

import (
	"verifiabledata/balloon"
	"verifiabledata/balloon/history"
)

type Snapshot struct {
	HyperDigest   []byte
	HistoryDigest []byte
	Version       uint
	Event         []byte
	//TODO: implement this
	// EventDigest   string
}

type HistoryNode struct {
	Digest       []byte
	Index, Layer uint
}

type Proofs struct {
	HyperAuditPath   [][]byte
	HistoryAuditPath []HistoryNode
}

type MembershipProof struct {
	Key                         []byte
	KeyDigest                   []byte
	IsMember                    bool
	Proofs                      *Proofs
	QueryVersion, ActualVersion uint
}

func ToSnapshot(commitment *balloon.Commitment, event string) *Snapshot {
	return &Snapshot{
		commitment.HyperDigest,
		commitment.HistoryDigest,
		commitment.Version,
		[]byte(event),
	}
}

func ToHistoryAuditPath(path []history.Node) []HistoryNode {
	result := make([]HistoryNode, 0)
	for _, elem := range path {
		result = append(result, HistoryNode{elem.Digest, elem.Index, elem.Layer})
	}
	return result
}

func ToMembershipProof(event string, proof *balloon.MembershipProof) *MembershipProof {
	return &MembershipProof{
		[]byte(event),
		proof.KeyDigest,
		proof.Exists,
		&Proofs{
			proof.HyperProof,
			ToHistoryAuditPath(proof.HistoryProof),
		},
		proof.QueryVersion,
		proof.ActualVersion,
	}
}

func ToHistoryNode(path []HistoryNode) []history.Node {
	result := make([]history.Node, 0)
	for _, elem := range path {
		result = append(result, history.Node{elem.Digest, elem.Index, elem.Layer})
	}
	return result
}
