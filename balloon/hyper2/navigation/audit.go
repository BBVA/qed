package navigation

import (
	"github.com/bbva/qed/hashing"
)

type AuditPath []hashing.Digest

func NewAuditPath() AuditPath {
	return make(AuditPath, 0)
}

func (p AuditPath) Get(index int) (hashing.Digest, bool) {
	if index >= len(p) {
		return nil, false
	}
	return p[index], true
}
