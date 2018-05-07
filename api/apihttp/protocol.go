package apihttp

import (
	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
)

type Event struct {
	Event []byte
}

type MembershipQuery struct {
	Key     []byte
	Version uint64
}

type Snapshot struct {
	HyperDigest   []byte
	HistoryDigest []byte
	Version       uint64
	Event         []byte
	//TODO: implement this
	// EventDigest   string
}

type HistoryNode struct {
	Digest       []byte
	Index, Layer uint64
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
	QueryVersion, ActualVersion uint64
}

func ToSnapshot(commitment *balloon.Commitment, event []byte) *Snapshot {
	return &Snapshot{
		commitment.HyperDigest,
		commitment.HistoryDigest,
		commitment.Version,
		event,
	}
}

func ToHistoryAuditPath(path []history.Node) []HistoryNode {
	result := make([]HistoryNode, 0)
	for _, elem := range path {
		result = append(result, HistoryNode{elem.Digest, elem.Index, elem.Layer})
	}
	return result
}

func ToMembershipProof(event []byte, proof *balloon.MembershipProof) *MembershipProof {
	return &MembershipProof{
		event,
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

func ToBalloonProof(id string, p *MembershipProof, hasher hashing.Hasher) *balloon.Proof {

	historyProof := history.NewProof(ToHistoryNode(p.Proofs.HistoryAuditPath), p.QueryVersion, hasher)
	hyperProof := hyper.NewProof(id, p.Proofs.HyperAuditPath, hasher)

	return balloon.NewProof(p.IsMember, hyperProof, historyProof, p.QueryVersion, p.ActualVersion, hasher)

}
