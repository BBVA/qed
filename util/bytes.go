package util

import "encoding/binary"

func UIntAsBytes(value uint) []byte {
	valueBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueBytes, uint32(value))
	return valueBytes
}

func UInt64AsBytes(value uint64) []byte {
	valueBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(valueBytes, uint64(value))
	return valueBytes
}
