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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lni/dragonboat"
	"github.com/lni/dragonboat-example/utils"
	"github.com/lni/dragonboat/config"
	"github.com/lni/dragonboat/logger"
)

func init() {
	//TODO: find a better place
	logger.GetLogger("raft").SetLevel(logger.ERROR)
	logger.GetLogger("rsm").SetLevel(logger.WARNING)
	logger.GetLogger("transport").SetLevel(logger.WARNING)
	logger.GetLogger("grpc").SetLevel(logger.WARNING)
}

// TODO: This is deprecated, remove when using a proper implementation
type RequestType uint64

const (
	exampleClusterID uint64 = 128
)

const (
	PUT RequestType = iota
	GET
)

var (
	// initial nodes count is fixed to three, their addresses are also fixed
	addresses = []string{
		"localhost:63001",
		"localhost:63002",
		"localhost:63003",
	}
)

func parseCommand(msg string) (RequestType, string, string, bool) {
	parts := strings.Split(strings.TrimSpace(msg), " ")
	if len(parts) == 0 || (parts[0] != "put" && parts[0] != "get") {
		return PUT, "", "", false
	}
	if parts[0] == "put" {
		if len(parts) != 3 {
			return PUT, "", "", false
		}
		return PUT, parts[1], parts[2], true
	}
	if len(parts) != 2 {
		return GET, "", "", false
	}
	return GET, parts[1], "", true
}

// TODO: EOF goto(TODO#1)

// WAL type exposes the contract to interact with the Write Ahead Log
type WAL struct {
	ch          chan string
	raftStopper *utils.Stopper // TODO: Remove this util dependecy
}

// NewWAL function is the constructor for WAL structs
func NewWAL(nodeID, clusterID uint64, nodeAddr string) WAL {

	// TODO: Integrate this with the outside
	peers := make(map[uint64]string)
	peers[nodeID] = nodeAddr
	join := false

	wal := WAL{
		ch:          make(chan string, 16),
		raftStopper: utils.NewStopper(),
	}

	rc := config.Config{
		NodeID:             nodeID,
		ClusterID:          exampleClusterID,
		ElectionRTT:        10,
		HeartbeatRTT:       1,
		CheckQuorum:        true,
		SnapshotEntries:    10,
		CompactionOverhead: 5,
	}

	datadir := filepath.Join("example-data", "helloworld-data", fmt.Sprintf("node%d", nodeID))

	nodeHostConfig := config.NodeHostConfig{
		WALDir:         datadir,
		NodeHostDir:    datadir,
		RTTMillisecond: 200,
		RaftAddress:    nodeAddr,
	}

	nodeHost, err := dragonboat.NewNodeHost(nodeHostConfig)
	if err != nil {
		panic(err)
	}

	// TODO: update NEWDiskKV to our own RSM
	err := nodeHost.StartOnDiskCluster(peers, join, NewDiskKV, rc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to add cluster, %v\n", err)
	}

	clientSession := nodeHost.GetNoOPSession(exampleClusterID)

	wal.raftStopper.RunWorker(func() {
		for {
			select {
			case v, ok := <-wal.ch:
				if !ok {
					return
				}
				msg := strings.Replace(v, "\n", "", 1)
				// input message must be in the following formats -
				// put key value
				// get key
				rt, key, val, ok := parseCommand(msg)
				if !ok {
					fmt.Fprintf(os.Stderr, "invalid input\n")
					continue
				}
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
				if rt == PUT {
					kv := &KVData{
						Key: key,
						Val: val,
					}
					data, err := json.Marshal(kv)
					if err != nil {
						panic(err)
					}
					_, err = nodeHost.SyncPropose(ctx, clientSession, data)
					if err != nil {
						fmt.Fprintf(os.Stderr, "SyncPropose returned error %v\n", err)
					}
				} else {
					result, err := nodeHost.SyncRead(ctx, exampleClusterID, []byte(key))
					if err != nil {
						fmt.Fprintf(os.Stderr, "SyncRead returned error %v\n", err)
					} else {
						fmt.Fprintf(os.Stdout, "query key: %s, result: %s\n", key, result)
					}
				}
				cancel()

			case <-wal.raftStopper.ShouldStop():
				return
			}
		}
	})

	wal.raftStopper.Wait()

	return wal

}

func (w *WAL) Close() {

}
