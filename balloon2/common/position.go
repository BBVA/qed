package common

type Position interface {
	Index() []byte
	Height() uint16
	Bytes() []byte
	String() string
	StringId() string
	IndexAsUint64() uint64
}
