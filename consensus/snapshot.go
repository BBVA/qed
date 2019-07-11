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

import (
	"github.com/bbva/qed/log"
	"github.com/hashicorp/go-msgpack/codec"
	"github.com/hashicorp/raft"
)

type fsmSnapshot struct {
	LastSeqNum     uint64
	BalloonVersion uint64
	ClusterInfo    []byte
}

// Persist writes the snapshot to the given sink.
func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Debug("Persisting snapshot with ID: %s", sink.ID())
	err := func() error {
		data, err := f.encode()
		if err != nil {
			return err
		}
		_, err = sink.Write(data)
		if err != nil {
			return err
		}
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

func (f *fsmSnapshot) encode() ([]byte, error) {
	var out []byte
	ch := new(codec.MsgpackHandle)
	enc := codec.NewEncoderBytes(&out, ch)
	err := enc.Encode(f)
	return out, err
}

func (f *fsmSnapshot) decode(in []byte) error {
	ch := new(codec.MsgpackHandle)
	return codec.NewDecoderBytes(in, ch).Decode(f)
}
