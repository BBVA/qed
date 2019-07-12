package consensus

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/bbva/qed/crypto/hashing"
	"github.com/bbva/qed/log"
	"github.com/bbva/qed/metrics"
	"github.com/bbva/qed/protocol"
	"github.com/bbva/qed/raftwal/raftrocks"
	"github.com/bbva/qed/storage"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
)

const (
	leaderWaitDelay = 100 * time.Millisecond
)

// ClusteringOptions contains node options related to clustering.
type ClusteringOptions struct {
	NodeID          string   // ID of the node within the cluster.
	Addr            string   // IP address where to listen for Raft commands.
	ClusterMgmtAddr string   // IP address where to listen for cluster GRPC operations.
	Bootstrap       bool     // Bootstrap the cluster as a seed node if there is no existing state.
	Peers           []string // List of cluster peer node IDs to bootstrap the cluster state.
	RaftLogPath     string   // Path to Raft log store directory.
	LogCacheSize    int      // Number of Raft log entries to cache in memory to reduce disk IO.
	LogSnapshots    int      // Number of Raft log snapshots to retain.
	TrailingLogs    int64    // Number of logs left after a snapshot.
	Sync            bool     // Do a file sync after every write to the Raft log and stable store.
	RaftLogging     bool     // Enable logging of Raft library (disabled by default since really verbose).

	// These will be set to some sane defaults. Change only if experiencing raft issues.
	RaftHeartbeatTimeout time.Duration
	RaftElectionTimeout  time.Duration
	RaftLeaseTimeout     time.Duration
	RaftCommitTimeout    time.Duration
}

func DefaultClusteringOptions() *ClusteringOptions {
	return &ClusteringOptions{
		NodeID:       "",
		Addr:         "",
		Bootstrap:    false,
		Peers:        make([]string, 0),
		RaftLogPath:  "",
		LogCacheSize: 512,
		LogSnapshots: 2,
		TrailingLogs: 10240,
		Sync:         false,
		RaftLogging:  false,
	}
}

type NodeInfo struct {
	NodeID          string
	RaftAddr        string
	ClusterMgmtAddr string
	HTTPAddr        string
	HTTPMgmtAddr    string
	MetricsAddr     string
}

type ClusterInfo struct {
	NodeID   string
	LeaderID string
	Nodes    map[string]NodeInfo
}

type RaftNode struct {
	path string
	info NodeInfo

	applyTimeout time.Duration

	db        storage.ManagedStore    // Persistent database
	raftLog   *raftrocks.RocksDBStore // Underlying rocksdb-backed persistent log store
	snapshots *raft.FileSnapshotStore // Persistent snapstop store

	raft       *raft.Raft             // The consensus mechanism
	transport  *raft.NetworkTransport // Raft network transport
	raftConfig *raft.Config           //Config provides any necessary configuration for the Raft server.
	grpcServer *grpc.Server

	fsm         *balloonFSM             // Balloon's finite state machine
	snapshotsCh chan *protocol.Snapshot // channel to publish snapshots

	hasherF         func() hashing.Hasher
	raftNodeMetrics *raftNodeMetrics // Raft node metrics.

	metrics *raftNodeMetrics

	sync.Mutex
	closed bool
}

func NewRaftNode(opts *ClusteringOptions, store storage.ManagedStore, snapshotsCh chan *protocol.Snapshot) (*RaftNode, error) {

	// We create s.raft early because once NewRaft() is called, the
	// raft code may asynchronously invoke FSM.Apply() and FSM.Restore()
	// So we want the object to exist so we can check on leader atomic, etc..
	node := &RaftNode{
		path: opts.RaftLogPath,
		info: NodeInfo{
			NodeID:          opts.NodeID,
			RaftAddr:        opts.Addr,
			ClusterMgmtAddr: opts.ClusterMgmtAddr,
		},
		snapshotsCh:  snapshotsCh,
		applyTimeout: 10 * time.Second,
	}

	// Create the log store
	raftLog, err := raftrocks.New(raftrocks.Options{
		Path:             opts.RaftLogPath + "/wal",
		NoSync:           !opts.Sync,
		EnableStatistics: true},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new Raft log: %s", err)
	}
	logStore, err := raft.NewLogCache(opts.LogCacheSize, raftLog)
	if err != nil {
		return nil, fmt.Errorf("cannot create a new cached log store: %s", err)
	}

	// Set hashing function
	hasherF := hashing.NewSha256Hasher

	// Instantiate balloon FSM
	fsm, err := newBalloonFSM(store, hasherF)
	if err != nil {
		return nil, fmt.Errorf("new balloon fsm: %s", err)
	}

	node.db = store
	node.raftLog = raftLog
	node.hasherF = hasherF
	node.fsm = fsm

	// setup Raft configuration
	conf := raft.DefaultConfig()
	if opts.RaftHeartbeatTimeout != 0 {
		conf.HeartbeatTimeout = opts.RaftHeartbeatTimeout
	}
	if opts.RaftHeartbeatTimeout != 0 {
		conf.ElectionTimeout = opts.RaftElectionTimeout
	}
	if opts.RaftHeartbeatTimeout != 0 {
		conf.LeaderLeaseTimeout = opts.RaftLeaseTimeout
	}
	if opts.RaftHeartbeatTimeout != 0 {
		conf.CommitTimeout = opts.RaftCommitTimeout
	}
	conf.LocalID = raft.ServerID(opts.NodeID)
	conf.Logger = hclog.Default()
	node.raftConfig = conf

	// setup Raft transport
	advertiseAddr, err := net.ResolveTCPAddr("tcp", node.info.RaftAddr)
	if err != nil {
		return nil, err
	}
	node.transport, err = raft.NewTCPTransportWithLogger(node.info.RaftAddr, advertiseAddr, 3, 10*time.Second, log.GetLogger())
	if err != nil {
		return nil, err
	}

	// create the snapshot store. This allows the Raft to truncate the log.
	// The library creates a folder to store the snapshots in.
	node.snapshots, err = raft.NewFileSnapshotStoreWithLogger(opts.RaftLogPath, opts.LogSnapshots, log.GetLogger())
	if err != nil {
		return nil, fmt.Errorf("file snapshot store: %s", err)
	}

	// instantiate the raft server
	node.raft, err = raft.NewRaft(node.raftConfig, node.fsm, logStore, node.raftLog, node.snapshots, node.transport)
	if err != nil {
		node.transport.Close()
		node.raftLog.Close()
		return nil, fmt.Errorf("new raft: %s", err)
	}

	// start grpc server to handle requests to join the cluster
	listener, err := net.Listen("tcp", node.info.ClusterMgmtAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %v", err)
	}
	node.grpcServer = grpc.NewServer()
	RegisterClusterServiceServer(node.grpcServer, node) // registers itself
	go node.grpcServer.Serve(listener)

	// register metrics
	node.metrics = newRaftNodeMetrics(node)

	// check existing state
	existingState, err := raft.HasExistingState(logStore, node.raftLog, node.snapshots)
	if err != nil {
		node.Shutdown(true)
		return nil, err
	}
	if existingState {
		log.Debugf("Loaded existing state for Raft.")
	} else {
		log.Debugf("No existing state found for Raft.")
		// Bootstrap if there is no previous state and we are starting this node as
		// a seed or a cluster configuration is provided.
		if opts.Bootstrap {
			log.Info("Bootstraping cluster...")
			if err := node.bootstrapCluster(); err != nil {
				node.Shutdown(true)
				return nil, err
			}
			log.Info("Cluster successfully bootstraped.")
		} else {
			log.Info("Attempting to join the cluster.")
			// Attempt to join the cluster if we're not bootstraping.
			err := node.AttemptToJoinCluster(opts.Peers)
			if err != nil {
				node.Shutdown(true)
				return nil, fmt.Errorf("failed to join Raft cluster")
			}
			log.Info("Join operation finished successfully.")
		}
	}

	return node, nil
}

func (n *RaftNode) bootstrapCluster() error {
	// include ourself in the cluster
	servers := []raft.Server{
		{
			ID:      n.raftConfig.LocalID,
			Address: n.transport.LocalAddr(),
		},
	}
	return n.raft.BootstrapCluster(raft.Configuration{Servers: servers}).Error()
}

func (n *RaftNode) Shutdown(wait bool) error {
	n.Lock()
	if n.closed {
		n.Unlock()
		return nil
	}
	defer func() {
		n.closed = true
		n.Unlock()
	}()

	// shutdown grpc
	if n.grpcServer != nil {
		n.grpcServer.GracefulStop()
		n.grpcServer = nil
	}

	// shutdown Raft
	if n.transport != nil {
		if err := n.transport.Close(); err != nil {
			return err
		}
		n.transport = nil
	}
	if n.raft != nil {
		f := n.raft.Shutdown()
		if wait {
			if e := f.(raft.Future); e.Error() != nil {
				return e.Error()
			}
		}
		n.raft = nil
	}
	if n.raftLog != nil {
		if err := n.raftLog.Close(); err != nil {
			return err
		}
		n.raftLog = nil
	}

	// close fsm
	if n.fsm != nil {
		n.fsm.Close()
		n.fsm = nil
	}

	// close the database
	if n.db != nil {
		if err := n.db.Close(); err != nil {
			return err
		}
		n.db = nil
	}

	return nil

}

func (n *RaftNode) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

// // WaitForLeader waits until the node becomes leader or time is out.
// func (n *RaftNode) WaitForLeader(timeout time.Duration) (string, error) {
// 	tck := time.NewTicker(leaderWaitDelay)
// 	defer tck.Stop()
// 	tmr := time.NewTimer(timeout)
// 	defer tmr.Stop()

// 	for {
// 		select {
// 		case <-tck.C:
// 			l := string(n.raft.Leader())
// 			if l != "" {
// 				return l, nil
// 			}
// 		case <-tmr.C:
// 			return "", fmt.Errorf("timeout expired")
// 		}
// 	}
// }

// JoinCluster joins a node, identified by id and located at addr, to this store.
// The node must be ready to respond to Raft communications at that address.
// This must be called from the Leader or it will fail.
func (n *RaftNode) JoinCluster(ctx context.Context, req *RaftJoinRequest) (*RaftJoinResponse, error) {

	// Drop the request if we're not the leader. There's no race condition
	// after this check because even if we proceed with the cluster add, it
	// will fail if the node is not the leader as cluster changes go
	// through the Raft log.
	if !n.IsLeader() {
		return nil, nil
	}

	log.Infof("received join request for remote node %s at %s", req.NodeId, req.RaftAddress)

	// Add the node as a voter. This is idempotent. No-op if the request
	// came from ourselves.
	f := n.raft.AddVoter(raft.ServerID(req.NodeId), raft.ServerAddress(req.RaftAddress), 0, 0)
	if err := f.Error(); err != nil {
		return nil, err
	}

	log.Infof("node %s at %s joined successfully", req.NodeId, req.RaftAddress)

	return &RaftJoinResponse{}, nil
}

func (n *RaftNode) AttemptToJoinCluster(addrs []string) error {
	for _, addr := range addrs {
		log.Debug("Joining existent Raft cluster in addr: ", addr)
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("failed to join node at %s: %s", addr, err.Error())
		}
		defer conn.Close()

		client := NewClusterServiceClient(conn)

		_, err = client.JoinCluster(context.Background(), &RaftJoinRequest{
			NodeId:      n.info.NodeID,
			RaftAddress: n.info.RaftAddr,
		})
		if err == nil {
			break
		}
	}
	return nil
}

// Info function returns Raft current node info.
func (n *RaftNode) Info() NodeInfo {
	return n.info
}

// ClusterInfo function returns Raft current node info plus certain raft cluster
// info. Used in /info/shard.
func (n *RaftNode) ClusterInfo() ClusterInfo {
	return ClusterInfo{} // TODO make metadata calls
}

// RegisterMetrics register raft metrics: prometheus collectors and raftLog metrics.
func (n *RaftNode) RegisterMetrics(registry metrics.Registry) {
	if registry != nil {
		n.raftLog.RegisterMetrics(registry)
	}
	registry.MustRegister(n.metrics.collectors()...)
}
