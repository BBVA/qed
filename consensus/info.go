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

	"github.com/bbva/qed/log"
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
	ci := new(ClusterInfo)
	ci.Nodes = make(map[string]*NodeInfo)
	leaderAddr := n.raft.Leader()

	configFuture := n.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		log.Infof("failed to get raft configuration: %v", err)
		return nil
	}

	for _, srv := range configFuture.Configuration().Servers {
		conn, err := grpc.Dial(string(srv.Address), grpc.WithInsecure())
		if err != nil {
			log.Infof("failed dialing grpc to %s: %v", string(srv.Address), err)
			continue
		}
		defer conn.Close()
		client := NewClusterServiceClient(conn)
		resp, err := client.FetchNodeInfo(context.Background(), new(InfoRequest))
		if err != nil {
			log.Infof("failed fetching info from %s: %v", string(srv.Address), err)
			continue
		}
		ci.Nodes[resp.NodeInfo.NodeId] = resp.NodeInfo
		if leaderAddr == srv.Address {
			ci.LeaderId = resp.NodeInfo.NodeId
		}
	}

	return ci
}
