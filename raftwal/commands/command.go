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

// Package commands defines the Raft log entries command types.
package commands

import (
	"bytes"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/hashicorp/go-msgpack/codec"
)

// CommandType are commands that affect the state of the cluster,
// and must go through raft.
type CommandType uint8

const (
	AddEventCommandType CommandType = iota // Commands which modify the database.
	AddEventsBulkCommandType
	MetadataSetCommandType
	MetadataDeleteCommandType
)

// AddEventCommand is used when inserting a single event digest.
type AddEventCommand struct {
	EventDigest hashing.Digest
}

// AddEventsBulkCommand is used when inserting a bulk of event digests.
type AddEventsBulkCommand struct {
	EventDigests []hashing.Digest
}

// MetadataSetCommand in used when adding metadata to a raft node.
type MetadataSetCommand struct {
	Id   string
	Data map[string]string
}

// MetadataDeleteCommand in used when deleting metadata from a raft node.
type MetadataDeleteCommand struct {
	Id string
}

// msgpackHandle is a shared handle for encoding/decoding of structs
var msgpackHandle = &codec.MsgpackHandle{}

// Decode is used to encode a MsgPack object with type prefix.
func Decode(buf []byte, out interface{}) error {
	return codec.NewDecoder(bytes.NewReader(buf), msgpackHandle).Decode(out)
}

// Encode is used to encode a MsgPack object with type prefix
func Encode(t CommandType, cmd interface{}) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte(uint8(t))
	err := codec.NewEncoder(&buf, msgpackHandle).Encode(cmd)
	return buf.Bytes(), err
}
