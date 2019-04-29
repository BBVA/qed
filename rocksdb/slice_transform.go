/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package rocksdb

// #include "rocksdb/c.h"
import "C"

// A SliceTransform is a generic pluggable way of transforming one string
// to another. Its primary use-case is in configuring rocksdb
//  to store prefix blooms by setting prefix_extractor in ColumnFamilyOptions.
type SliceTransform interface {
	// Transform extracts a prefix from a specified key. This method is called
	// when a key is inserted into the db, and the returned slice is used to
	// create a bloom filter.
	Transform(key []byte) []byte

	// InDomain determines whether the specified key is compatible with the
	// logic specified in the Transform method. This method is invoked for
	// every key that is inserted into the db. If this method returns true,
	// then Transform is called to translate the key to its prefix and
	// that returned prefix is inserted into the bloom filter. If this
	// method returns false, then the call to Transform is skipped and
	// no prefix is inserted into the bloom filters.
	//
	// For example, if the Transform method operates on a fixed length
	// prefix of size 4, then an invocation to InDomain("abc") returns
	// false because the specified key length(3) is shorter than the
	// prefix size of 4.
	//
	// Wiki documentation here:
	// https://github.com/facebook/rocksdb/wiki/Prefix-Seek-API-Changes
	//
	InDomain(key []byte) bool

	// InRange is currently not used and remains here for backward compatibility.
	InRange(key []byte) bool

	// Name returns the name of this transformation.
	Name() string
}

// NewFixedPrefixTransform creates a new fixed prefix transform.
func NewFixedPrefixTransform(prefixLen int) SliceTransform {
	return NewNativeSliceTransform(C.rocksdb_slicetransform_create_fixed_prefix(C.size_t(prefixLen)))
}

// NewNativeSliceTransform creates a SliceTransform object.
func NewNativeSliceTransform(c *C.rocksdb_slicetransform_t) SliceTransform {
	return nativeSliceTransform{c}
}

type nativeSliceTransform struct {
	c *C.rocksdb_slicetransform_t
}

func (st nativeSliceTransform) Transform(src []byte) []byte { return nil }
func (st nativeSliceTransform) InDomain(src []byte) bool    { return false }
func (st nativeSliceTransform) InRange(src []byte) bool     { return false }
func (st nativeSliceTransform) Name() string                { return "" }

// Hold references to slice transforms.
var sliceTransforms = NewCOWList()

type sliceTransformWrapper struct {
	name           *C.char
	sliceTransform SliceTransform
}

func registerSliceTransform(st SliceTransform) int {
	return sliceTransforms.Append(sliceTransformWrapper{C.CString(st.Name()), st})
}

//export rocksdb_slicetransform_transform
func rocksdb_slicetransform_transform(idx int, cKey *C.char, cKeyLen C.size_t, cDstLen *C.size_t) *C.char {
	key := charToBytes(cKey, cKeyLen)
	dst := sliceTransforms.Get(idx).(sliceTransformWrapper).sliceTransform.Transform(key)
	*cDstLen = C.size_t(len(dst))
	return cByteSlice(dst)
}

//export rocksdb_slicetransform_in_domain
func rocksdb_slicetransform_in_domain(idx int, cKey *C.char, cKeyLen C.size_t) C.uchar {
	key := charToBytes(cKey, cKeyLen)
	inDomain := sliceTransforms.Get(idx).(sliceTransformWrapper).sliceTransform.InDomain(key)
	return boolToUchar(inDomain)
}

//export rocksdb_slicetransform_in_range
func rocksdb_slicetransform_in_range(idx int, cKey *C.char, cKeyLen C.size_t) C.uchar {
	key := charToBytes(cKey, cKeyLen)
	inRange := sliceTransforms.Get(idx).(sliceTransformWrapper).sliceTransform.InRange(key)
	return boolToUchar(inRange)
}

//export rocksdb_slicetransform_name
func rocksdb_slicetransform_name(idx int) *C.char {
	return sliceTransforms.Get(idx).(sliceTransformWrapper).name
}
