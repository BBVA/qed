package common

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
