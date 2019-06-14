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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageEncodeDecode(t *testing.T) {
	var m2 Message
	m1 := &Message{
		Kind:    BatchMessageType,
		From:    nil,
		TTL:     0,
		Payload: nil,
	}

	buff, err := m1.Encode()
	require.NoError(t, err, "Encoding must end succesfully")
	_ = m2.Decode(buff)
	require.Equal(t, &m2, m1, "Messages must be equal")

}
