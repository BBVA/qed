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

package history

import (
	"fmt"

	"github.com/bbva/qed/util"
)

type HistoryPosition struct {
	index  uint64
	height uint16
}

func NewPosition(index uint64, height uint16) *HistoryPosition {
	return &HistoryPosition{
		index:  index,
		height: height,
	}
}

func (p HistoryPosition) Index() []byte {
	return util.Uint64AsBytes(p.index)
}

func (p HistoryPosition) Height() uint16 {
	return p.height
}

func (p HistoryPosition) IndexAsUint64() uint64 {
	return p.index
}

func (p HistoryPosition) Bytes() []byte {
	b := make([]byte, 10) // Size of the index plus 2 bytes for the height
	indexAsBytes := p.Index()
	copy(b, indexAsBytes)
	copy(b[len(indexAsBytes):], util.Uint16AsBytes(p.height))
	return b
}

func (p HistoryPosition) String() string {
	return fmt.Sprintf("Pos(%d, %d)", p.IndexAsUint64(), p.height)
}

func (p HistoryPosition) StringId() string {
	return fmt.Sprintf("%d|%d", p.IndexAsUint64(), p.height)
}
