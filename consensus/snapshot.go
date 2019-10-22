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
	"context"
	"errors"
	io "io"

	"github.com/bbva/qed/storage"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type fsmSnapshot struct {
	LastSeqNum     uint64
	BalloonVersion uint64
}

// Persist writes the snapshot to the given sink.
func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
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
func (f *fsmSnapshot) Release() {}

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
	if err := c.srv.Send(chunk); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (c chunkWriter) Close() error {
	return nil
}

type chunkReader struct {
	conn   *grpc.ClientConn
	stream ClusterService_FetchSnapshotClient
	buf    *bytes.Buffer
	done   bool
}

func (c *chunkReader) Read(p []byte) (int, error) {
	var n int
	var err error
	for {
		if c.buf.Len() == 0 {
			chunk, err := c.stream.Recv()
			if err != nil {
				return 0, err
			}
			c.buf.Write(chunk.Content)
		}
		n, err = c.buf.Read(p)
		if err == nil {
			break
		}
		c.buf.Reset()
	}
	return n, err
}

func (c *chunkReader) Close() error {
	if err := c.stream.CloseSend(); err != nil {
		return err
	}
	return c.conn.Close()
}

func newChunkReader(conn *grpc.ClientConn, stream ClusterService_FetchSnapshotClient) *chunkReader {
	cr := new(chunkReader)
	cr.conn = conn
	cr.stream = stream
	cr.buf = new(bytes.Buffer)
	cr.done = true
	return cr
}

func (n *RaftNode) FetchSnapshot(req *FetchSnapshotRequest, srv ClusterService_FetchSnapshotServer) error {
	chunker := &chunkWriter{srv: srv}

	validateF := func(lastAppliedVersion uint64) storage.ValidateF {
		lastSnapshotAppliedVersion := lastAppliedVersion
		return func(meta []byte) (bool, error) {
			metadata := new(VersionMetadata)
			err := decodeMsgPack(meta, metadata)
			if err != nil {
				return false, nil
			}
			if metadata.PreviousVersion > lastSnapshotAppliedVersion {
				return false, errors.New("Gap found between versions")
			}
			if metadata.NewVersion < lastSnapshotAppliedVersion {
				// apply only those who are ahead the version specified with the parameter.
				return false, nil
			}
			if metadata.NewVersion == lastSnapshotAppliedVersion && lastSnapshotAppliedVersion != 0 {
				return false, nil
			}
			lastSnapshotAppliedVersion = metadata.NewVersion
			return true, nil
		}
	}

	return n.db.FetchSnapshot(chunker, req.StartSeqNum, req.EndSeqNum, validateF(req.LastAppliedVersion))
}

func (n *RaftNode) attemptToFetchSnapshot(lastSeqNum, lastAppliedVersion uint64) (io.ReadCloser, error) {
	leaderAddr := string(n.raft.Leader())
	conf, err := n.tlsConfigurator.OutgoingTLSConfig()
	if err != nil {
		return nil, err
	}
	var conn *grpc.ClientConn
	if conf != nil {
		conn, err = grpc.Dial(leaderAddr, grpc.WithTransportCredentials(credentials.NewTLS(conf)))
	} else {
		conn, err = grpc.Dial(leaderAddr, grpc.WithInsecure())
	}
	if err != nil {
		return nil, err
	}
	client := NewClusterServiceClient(conn)
	stream, err := client.FetchSnapshot(context.Background(), &FetchSnapshotRequest{
		LastAppliedVersion: lastAppliedVersion,
		StartSeqNum:        n.db.LastWALSequenceNumber(),
		EndSeqNum:          lastSeqNum})
	if err != nil {
		return nil, err
	}

	return newChunkReader(conn, stream), nil
}
