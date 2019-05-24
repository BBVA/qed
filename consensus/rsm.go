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

package consensus

/*
type IStateMachine interface {
	Open(<-chan struct{}) (uint64, error)
	Update(entries []sm.Entry) ([]sm.Entry, error)
	Lookup(query []byte) ([]byte, error)
	Sync() error
	PrepareSnapshot() (interface{}, error)
	SaveSnapshot(interface{},
		io.Writer, sm.ISnapshotFileCollection, <-chan struct{}) error
	RecoverFromSnapshot(uint64, io.Reader, []sm.SnapshotFile, <-chan struct{}) error
	Close()
	GetHash() (uint64, error)
	ConcurrentSnapshot() bool
	OnDiskStateMachine() bool
	StateMachineType() pb.StateMachineType
}
*/
