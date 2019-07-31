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

	"github.com/bbva/qed/log"
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

	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ci := new(ClusterInfo)

	done := make(chan struct{})

	go func() {
		var lock sync.Mutex
		nodes := make(map[string]*NodeInfo)
		leaderAddr := string(n.raft.Leader())
		if leaderAddr == "" {
			leaderAddr = "unknown"
		}

		servers := listServers(ctx, n.raft)

		for _, srv := range servers {
			wg.Add(1)
			go func(addr string) {
				defer wg.Done()
				resp, err := grpcFetchInfo(ctx, addr)
				if err != nil {
					log.Infof("Error geting node info from %s: %v", addr, err)
					return
				}
				lock.Lock()
				if leaderAddr == addr {
					ci.LeaderId = resp.NodeInfo.NodeId
				}
				nodes[resp.NodeInfo.NodeId] = resp.NodeInfo
				lock.Unlock()
			}(string(srv.Address))
		}
		wg.Wait()
		ci.Nodes = nodes
		close(done)
	}()

	select {
	case <-ctx.Done():
		log.Infof("Timed out  geting cluster  info ")
		return nil
	case <-done:
		return ci
	}
}

func listServers(ctx context.Context, r *raft.Raft) []raft.Server {
	var list []raft.Server
	done := make(chan struct{})

	go func() {
		configFuture := r.GetConfiguration()
		if err := configFuture.Error(); err != nil {
			log.Infof("Error getting configuration from raft: %v", err)
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
