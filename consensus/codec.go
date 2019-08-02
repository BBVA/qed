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
	"github.com/hashicorp/go-msgpack/codec"
)

var handler *codec.MsgpackHandle

func init() {
	handler = new(codec.MsgpackHandle)
}

// Decode reverses the encode operation on a byte slice input
func decodeMsgPack(buf []byte, out interface{}) error {
	dec := codec.NewDecoderBytes(buf, handler)
	return dec.Decode(out)
}

// Encode writes an encoded object to a new bytes buffer
func encodeMsgPack(in interface{}) ([]byte, error) {
	var buf []byte
	dec := codec.NewEncoderBytes(&buf, handler)
	err := dec.Encode(in)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
