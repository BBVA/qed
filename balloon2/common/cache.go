package common

type Cache interface {
	Get(pos Position) (Digest, bool)
}

type ModifiableCache interface {
	Put(pos Position, value Digest)
	Cache
}
