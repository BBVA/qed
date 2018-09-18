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

import (
	"encoding/binary"
	"fmt"

	"github.com/bbva/qed/util"
)

type HyperPosition struct {
	index  []byte
	height uint16
}

func NewPosition(index []byte, height uint16) *HyperPosition {
	return &HyperPosition{
		index:  index,
		height: height,
	}
}

func (p HyperPosition) Index() []byte {
	return p.index
}

func (p HyperPosition) Height() uint16 {
	return p.height
}

func (p HyperPosition) Bytes() []byte {
	b := make([]byte, 34) // Size of the index plus 2 bytes for the height
	copy(b, p.index)
	copy(b[len(p.index):], util.Uint16AsBytes(p.height))
	return b
}

func (p HyperPosition) String() string {
	return fmt.Sprintf("Pos(%x, %d)", p.index, p.height)
}

func (p HyperPosition) StringId() string {
	return fmt.Sprintf("%x|%d", p.index, p.height)
}

func (p HyperPosition) IndexAsUint64() uint64 {
	return binary.LittleEndian.Uint64(p.index)
}
