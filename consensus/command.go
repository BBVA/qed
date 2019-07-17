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
	"bytes"
	"fmt"
)

// commandType are commands that affect the state of the cluster,
// and must go through raft.
type commandType uint8

const (
	addEventCommandType commandType = iota // Commands which modify the database.
	infoSetCommandType
)

type command struct {
	id   commandType
	data []byte
}

func (c *command) encode(in interface{}) error {
	var buf bytes.Buffer
	buf.WriteByte(uint8(c.id))
	data, err := encodeMsgPack(in)
	if err != nil {
		return err
	}
	_, err = buf.Write(data)
	if err != nil {
		return err
	}
	c.data = buf.Bytes()
	return nil
}

func (c *command) decode(out interface{}) error {
	if c.data == nil {
		return fmt.Errorf("Command is empty")
	}
	if c.id != commandType(c.data[0]) {
		return fmt.Errorf("Command type %v is not %v", c.id, commandType(c.data[0]))
	}
	return decodeMsgPack(c.data[1:], out)
}

func newCommand(t commandType) *command {
	return &command{
		id: t,
	}
}

func newCommandFromRaft(data []byte) *command {
	return &command{
		id:   commandType(data[0]),
		data: data,
	}
}
