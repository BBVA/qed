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

func TestEncodeDecode(t *testing.T) {
	var m1, m2 Meta

	m1.Role = "string test"

	buff, err := m1.Encode()
	require.NoError(t, err, "Error encoding metadata")
	err = m2.Decode(buff)
	require.NoError(t, err, "Error decoding metadata")
	require.Equal(t, m1, m2, "Both metadata must be equals")
}
