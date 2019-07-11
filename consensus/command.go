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
	"github.com/bbva/qed/crypto/hashing"
)

// commandType are commands that affect the state of the cluster,
// and must go through raft.
type commandType uint8

const (
	addEventCommandType commandType = iota // Commands which modify the database.
	addEventsBulkCommandType
	metadataSetCommandType
	metadataDeleteCommandType
)

// addEventCommand is used when inserting a single event digest.
type addEventCommand struct {
	EventDigest hashing.Digest
}

// AddEventsBulkCommand is used when inserting a bulk of event digests.
type addEventsBulkCommand struct {
	EventDigests []hashing.Digest
}

// MetadataSetCommand in used when adding metadata to a raft node.
type metadataSetCommand struct {
	ID   string
	Data map[string]*NodeInfo
}

// MetadataDeleteCommand in used when deleting metadata from a raft node.
type metadataDeleteCommand struct {
	ID string
}
