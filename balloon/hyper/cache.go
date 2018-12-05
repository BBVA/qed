/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

package hyper

import "github.com/bbva/qed/balloon/navigator"

type CacheResolver interface {
	ShouldBeInCache(pos navigator.Position) bool
	ShouldCache(pos navigator.Position) bool
	ShouldCollect(pos navigator.Position) bool
	IsOnPath(pos navigator.Position) bool
}

type SingleTargetedCacheResolver struct {
	numBits    uint16
	cacheLevel uint16
	targetKey  []byte
}

func NewSingleTargetedCacheResolver(numBits, cacheLevel uint16, targetKey []byte) *SingleTargetedCacheResolver {
	return &SingleTargetedCacheResolver{numBits, cacheLevel, targetKey}
}

func (r SingleTargetedCacheResolver) ShouldBeInCache(pos navigator.Position) bool {
	return pos.Height() >= r.cacheLevel && !r.IsOnPath(pos)
}

func (r SingleTargetedCacheResolver) ShouldCache(pos navigator.Position) bool {
	return pos.Height() == r.cacheLevel
}

func (r SingleTargetedCacheResolver) ShouldCollect(pos navigator.Position) bool {
	return pos.Height() == r.cacheLevel
}

/*
	This method does not reach leafs. Goes from root (bit := 0) to height=1 (bit := numbits - 1)
*/
func (r SingleTargetedCacheResolver) IsOnPath(pos navigator.Position) bool {
	height := pos.Height()
	if height == r.numBits {
		return true
	}
	bit := r.numBits - height - 1
	return bitIsSet(r.targetKey, bit) == bitIsSet(pos.Index(), bit)
}

/*
	Is bit in position 'i' set to 1?
	i   :	 2				3
	bits: [00101011]	[00101011]
	mask: [00100000]	[00010000]
			 true			false
*/
func bitIsSet(bits []byte, i uint16) bool {
	return bits[i/8]&(1<<uint(7-i%8)) != 0
}
