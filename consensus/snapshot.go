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
	"fmt"
	io "io"

	"github.com/bbva/qed/log"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
)

type fsmSnapshot struct {
	LastSeqNum     uint64
	BalloonVersion uint64
	Info           *ClusterInfo
}

// Persist writes the snapshot to the given sink.
func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Debug("Persisting snapshot with ID: %s", sink.ID())
	err := func() error {
		data, err := f.encode()
		if err != nil {
			return err
		}
		_, err = sink.Write(data)
		if err != nil {
			return err
		}
		return sink.Close()
	}()
	if err != nil {
		_ = sink.Cancel()
	}
	return err
}

// Release is invoked when we are finished with the snapshot.
func (f *fsmSnapshot) Release() {
	log.Debug("Snapshot created.")
}

func (f *fsmSnapshot) encode() ([]byte, error) {
	return encodeMsgPack(f)
}

func (f *fsmSnapshot) decode(in []byte) error {
	return decodeMsgPack(in, f)
}

type chunkWriter struct {
	srv ClusterService_FetchSnapshotServer
}

func (c chunkWriter) Write(data []byte) (int, error) {
	chunk := new(Chunk)
	chunk.Content = data
	fmt.Printf("chunkWriter: %v\n", data)
	if err := c.srv.Send(chunk); err != nil {
		return 0, err
	}
	return len(data), nil
}

type chunkReader struct {
	stream ClusterService_FetchSnapshotClient
}

func (c chunkReader) Read(p []byte) (int, error) {
	chunk, err := c.stream.Recv()
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	fmt.Printf("chunkReader: %v\n", chunk.Content)
	copy(p, chunk.Content)
	//p = append(p, chunk.Content...)
	return len(p), nil
}

func (n *RaftNode) FetchSnapshot(req *FetchSnapshotRequest, srv ClusterService_FetchSnapshotServer) error {
	chunker := &chunkWriter{srv: srv}
	return n.db.FetchSnapshot(chunker, req.SeqNum, n.db.LastWALSequenceNumber())
}

func (n *RaftNode) attemptToFetchSnapshot(seqNum uint64) (io.Reader, error) {

	nodeInfo := n.clusterInfo.Nodes[n.clusterInfo.LeaderId]
	conn, err := grpc.Dial(nodeInfo.ClusterMgmtAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := NewClusterServiceClient(conn)
	stream, err := client.FetchSnapshot(context.Background(), &FetchSnapshotRequest{seqNum})
	if err != nil {
		return nil, err
	}

	return &chunkReader{stream: stream}, nil
}
