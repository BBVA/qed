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
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
)

func (n *RaftNode) FetchNodeInfo(ctx context.Context, req *InfoRequest) (*InfoResponse, error) {
	resp := new(InfoResponse)
	resp.NodeInfo = n.info
	return resp, nil
}

// Info function returns Raft current node info.
func (n *RaftNode) Info() *NodeInfo {
	return n.info
}

// ClusterInfo function returns Raft current node info plus certain raft cluster
// info. Used in /info/shard.
func (n *RaftNode) ClusterInfo() *ClusterInfo {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ci := new(ClusterInfo)
	ci.Nodes = make(map[string]*NodeInfo)

	leaderAddr := string(n.raft.Leader())
	if leaderAddr == "" {
		return ci
	}

	servers := n.listServers(ctx, n.raft)

	var wg sync.WaitGroup
	infoCh := make(chan *NodeInfo, len(servers))
	done := make(chan struct{})

	wg.Add(len(servers))
	for _, srv := range servers {

		go func(addr string) {
			defer wg.Done()
			resp, err := grpcFetchInfo(ctx, addr)
			if err != nil {
				n.log.Infof("Error getting node info from %s: %v", addr, err)
				return
			}
			infoCh <- resp.NodeInfo
		}(string(srv.Address))
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for {
		select {
		case <-done:
			return ci
		case node := <-infoCh:
			if leaderAddr == node.RaftAddr {
				ci.LeaderId = node.NodeId
			}
			ci.Nodes[node.NodeId] = node
		}
	}

}

func (n *RaftNode) listServers(ctx context.Context, r *raft.Raft) []raft.Server {
	var list []raft.Server
	done := make(chan struct{})

	go func() {
		configFuture := r.GetConfiguration()
		if err := configFuture.Error(); err != nil {
			n.log.Infof("Error getting configuration from raft: %v", err)
		}
		list = configFuture.Configuration().Servers
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-done:
		return list
	}

}

func grpcFetchInfo(ctx context.Context, addr string) (*InfoResponse, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := NewClusterServiceClient(conn)
	resp, err := client.FetchNodeInfo(ctx, new(InfoRequest))
	if err != nil {
		return nil, err
	}
	return resp, nil
}
