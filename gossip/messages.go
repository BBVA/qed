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

package gossip

import (
	"bytes"

	"github.com/hashicorp/go-msgpack/codec"
)

// msgpackHandle is a shared handle for encoding/decoding of structs
var msgpackHandle = &codec.MsgpackHandle{}

const (
	MAXMESSAGEID = 1 << 8
)

type MessageType uint8

const (
	BatchMessageType MessageType = iota // Contains a protocol.BatchSnapshots
)

// Gossip message code. Up to 255 different messages.
type Message struct {
	Kind    MessageType
	From    *Peer
	TTL     int
	Payload []byte
}

/*
func (m *Message) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Decode(msg []byte) error {
	err := json.Unmarshal(msg, m)
	return err
}
*/

func (m *Message) Encode() ([]byte, error) {
	var buf bytes.Buffer
	err := codec.NewEncoder(&buf, msgpackHandle).Encode(m)
	return buf.Bytes(), err
}

func (m *Message) Decode(buf []byte) error {
	return codec.NewDecoder(bytes.NewReader(buf), msgpackHandle).Decode(m)
}

