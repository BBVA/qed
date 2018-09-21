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
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bbva/qed/storage/badger"
	"github.com/stretchr/testify/require"
)

var raftPath, badgerPath string

func init() {
	badgerPath = "/var/tmp/raft-test/badger"
	raftPath = "/var/tmp/raft-test/raft"
}

func raftAddr(id int) string {
	return fmt.Sprintf("127.0.0.1:830%d", id)
}
func joinAddr(id int) string {
	return fmt.Sprintf("127.0.0.1:840%d", id)
}

func Test_IsLeader(t *testing.T) {
	// t.Parallel()

	os.MkdirAll(badgerPath, os.FileMode(0755))
	badger, err := badger.NewBadgerStore(badgerPath)
	require.NoError(t, err)

	os.MkdirAll(raftPath, os.FileMode(0755))
	r, err := NewRaftBalloon(raftPath, raftAddr(0), "0", badger)
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(badgerPath)
		os.RemoveAll(raftPath)
	}()

	err = r.Open("", raftAddr(0), "0")
	require.NoError(t, err)

	defer r.Close(true)
	_, err = r.WaitForLeader(10 * time.Second)
	require.NoError(t, err)

	require.True(t, r.IsLeader(), "single node is not leader!")

}
