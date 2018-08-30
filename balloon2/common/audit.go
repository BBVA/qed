package common

type AuditPath map[string]Digest

func (p AuditPath) Get(pos Position) (Digest, bool) {
	digest, ok := p[pos.StringId()]
	return digest, ok
}
