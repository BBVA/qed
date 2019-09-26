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
	"github.com/pkg/errors"
)

// Agent metadata
type Meta struct {
	Role string
}

func (a *Meta) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := codec.NewEncoder(&buf, &codec.MsgpackHandle{})
	if err := encoder.Encode(a); err != nil {
		return nil, errors.Wrap(err, "Failed to encode agent metadata")
	}
	return buf.Bytes(), nil
}

func (a *Meta) Decode(buf []byte) error {
	reader := bytes.NewReader(buf)
	decoder := codec.NewDecoder(reader, &codec.MsgpackHandle{})
	if err := decoder.Decode(a); err != nil {
		return errors.Wrap(err, "Failed to decode agent metadata")
	}
	return nil
}
