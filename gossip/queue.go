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
package gossip

import (
	"runtime"
	"sync/atomic"

	"github.com/bbva/qed/protocol"
)

const queueSize uint64 = 65535

// Masking is faster than division
const indexMask uint64 = queueSize - 1

type Queue struct {
	// The padding members 1 to 5 below are here to ensure each item is on a separate cache line.
	// This prevents false sharing and hence improves performance.
	padding1           [8]uint64
	lastCommittedIndex uint64
	padding2           [8]uint64
	nextFreeIndex      uint64
	padding3           [8]uint64
	readerIndex        uint64
	padding4           [8]uint64
	contents           [queueSize]*protocol.BatchSnapshots
	padding5           [8]uint64
}

func NewQueue() *Queue {
	return &Queue{
		lastCommittedIndex: 0,
		nextFreeIndex:      1,
		readerIndex:        1,
	}
}

func (self *Queue) Write(value *protocol.BatchSnapshots) {
	var myIndex = atomic.AddUint64(&self.nextFreeIndex, 1) - 1
	//Wait for reader to catch up, so we don't clobber a slot which it is (or will be) reading
	for myIndex > (self.readerIndex + queueSize - 2) {
		runtime.Gosched()
	}
	//Write the item into it's slot
	self.contents[myIndex&indexMask] = value
	//Increment the lastCommittedIndex so the item is available for reading
	for !atomic.CompareAndSwapUint64(&self.lastCommittedIndex, myIndex-1, myIndex) {
		runtime.Gosched()
	}
}

func (self *Queue) Read() *protocol.BatchSnapshots {
	var myIndex = atomic.AddUint64(&self.readerIndex, 1) - 1
	//If reader has out-run writer, wait for a value to be committed
	for myIndex > self.lastCommittedIndex {
		runtime.Gosched()
	}
	return self.contents[myIndex&indexMask]
}
