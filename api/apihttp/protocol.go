package apihttp

import (
	"fmt"
	"verifiabledata/balloon"
	"verifiabledata/balloon/history"
)

type HistoryNode struct {
	Digest       string
	Index, Layer uint
}

type Proofs struct {
	HyperAuditPath   []string
	HistoryAuditPath []HistoryNode
}

type MembershipProof struct {
	Key                         string
	IsMember                    bool
	Proofs                      *Proofs
	QueryVersion, ActualVersion uint
}

func assemblyHyperAuditPath(path [][]byte) []string {
	result := make([]string, 0)
	for _, elem := range path {
		result = append(result, fmt.Sprintf("%x", elem))
	}
	return result
}

func assemblyHistoryAuditPath(path []history.Node) []HistoryNode {
	result := make([]HistoryNode, 0)
	for _, elem := range path {
		result = append(result, assemblyHistoryNode(elem))
	}
	return result
}

func assemblyHistoryNode(node history.Node) HistoryNode {
	return HistoryNode{
		fmt.Sprintf("%x", node.Digest),
		node.Index,
		node.Layer,
	}
}

func assemblyMembershipProof(event string, proof *balloon.MembershipProof) *MembershipProof {
	return &MembershipProof{
		event,
		proof.Exists,
		&Proofs{
			assemblyHyperAuditPath(proof.HyperProof),
			assemblyHistoryAuditPath(proof.HistoryProof),
		},
		proof.QueryVersion,
		proof.ActualVersion,
	}
}
