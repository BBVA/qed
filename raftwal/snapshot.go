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

package raftwal

import (
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
	"github.com/hashicorp/raft"
)

type fsmSnapshot struct {
	id    uint64
	store storage.ManagedStore
	meta  []byte
}

// Persist writes the snapshot to the given sink.
func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Debug("Persisting snapshot...")
	err := func() error {
		// if err := f.store.Dump(sink, f.id); err != nil {
		// 	return err
		// }
		return sink.Close()
	}()
	if err != nil {
		_ = sink.Cancel()
	}
	return err
}

// Release is invoked when we are finished with the snapshot.
func (f *fsmSnapshot) Release() {
	log.Debug("Snapshot created.")
}
