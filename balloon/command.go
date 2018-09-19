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

package balloon

import (
	"encoding/json"
)

// commandType are commands that affect the state of the cluster,
// and must go through raft.
type commandType int

const (
	insert commandType = iota // Commands which modify the database.
	//query                     // Commands which query the database.
)

type command struct {
	Type commandType     `json:"type,omitempty"`
	Sub  json.RawMessage `json:"sub,omitempty"`
}

func newCommand(t commandType, d interface{}) (*command, error) {
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	return &command{
		Type: t,
		Sub:  b,
	}, nil
}

type insertSubCommand struct {
	Event              []byte `json:"event,omitempty"`
	LastBalloonVersion uint64 `json:"event,omitempty`
}

func newInsertSubCommand(event []byte, lastBalloonVersion uint64) *insertSubCommand {
	return &insertSubCommand{event, lastBalloonVersion}
}
