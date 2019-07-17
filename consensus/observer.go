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
	"fmt"

	"github.com/hashicorp/raft"
)

var observationsFilterFn = func(o *raft.Observation) bool {
	switch o.Data.(type) {
	case raft.PeerObservation:
		return true
	case raft.LeaderObservation:
		return true
	default:
		return false
	}
}

func (n *RaftNode) startObservationsConsumer() {
	for {
		select {
		case obs := <-n.observationsCh:
			switch obs.Data.(type) {
			case raft.PeerObservation:
				peerObs := obs.Data.(raft.PeerObservation)
				if peerObs.Removed {
					n.infoMu.Lock()
					delete(n.clusterInfo.Nodes, string(peerObs.Peer.ID))
					n.infoMu.Unlock()
					cmd := newCommand(infoSetCommandType)
					cmd.encode(n.clusterInfo)
					n.propose(cmd)
				} 		
				fmt.Printf("ID[%s] - %+v\n", n.info.NodeId, peerObs)
			case raft.LeaderObservation:
				fmt.Printf("ID[%s] - %+v\n", n.info.NodeId, obs.Data)
				id, err := n.leaderID()
				if err == nil {
					n.infoMu.Lock()
					n.clusterInfo.LeaderId = id
					n.infoMu.Unlock()
				}
			default:
			}
		case <-n.done:
			return
		}
	}
}
