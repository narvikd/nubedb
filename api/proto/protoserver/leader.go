package protoserver

import (
	"context"
	"errors"
	"github.com/hashicorp/raft"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster"
)

// ExecuteOnLeader executes a command on the Raft leader.
//
// The leader is the node that is currently responsible for coordinating updates to the network.
func (srv *server) ExecuteOnLeader(ctx context.Context, req *proto.ExecuteOnLeaderRequest) (*proto.Empty, error) {
	log.Println("[proto] (ExecuteOnLeader) request received, processing...")

	// Applies the command to the leader
	errExecute := cluster.ApplyLeaderFuture(srv.Node.Consensus, req.Payload)
	if errExecute != nil {
		return &proto.Empty{}, errExecute
	}

	log.Println("[proto] (ExecuteOnLeader) request successful")
	return &proto.Empty{}, nil
}

// IsLeader checks if the node is currently the Raft leader.
func (srv *server) IsLeader(ctx context.Context, req *proto.Empty) (*proto.IsLeaderResponse, error) {
	log.Println("[proto] (IsLeader) request received, processing...")
	is := srv.Node.Consensus.State() == raft.Leader
	log.Println("[proto] (IsLeader) request successful")
	return &proto.IsLeaderResponse{IsLeader: is}, nil
}

// ConsensusJoin adds a new node to the Raft consensus network.
func (srv *server) ConsensusJoin(ctx context.Context, req *proto.ConsensusRequest) (*proto.Empty, error) {
	log.Println("[proto] (ConsensusJoin) request received, processing...")

	// Checks if the node is already part of the network
	consensusCfg := srv.Node.Consensus.GetConfiguration().Configuration()
	for _, s := range consensusCfg.Servers {
		if req.NodeID == string(s.ID) {
			return &proto.Empty{}, errors.New("node was already part of the network")
		}
	}

	// Add the new node to the network
	future := srv.Node.Consensus.AddVoter(raft.ServerID(req.NodeID), raft.ServerAddress(req.NodeConsensusAddr), 0, 0)
	if future.Error() != nil {
		return &proto.Empty{}, future.Error()
	}

	log.Println("[proto] (ConsensusJoin) request successful")
	return &proto.Empty{}, nil
}

// ConsensusRemove removes a node from the Raft consensus network.
func (srv *server) ConsensusRemove(ctx context.Context, req *proto.ConsensusRequest) (*proto.Empty, error) {
	log.Println("[proto] (ConsensusRemove) request received, processing...")

	// Removes the node from the network
	future := srv.Node.Consensus.RemoveServer(raft.ServerID(req.NodeID), 0, 0)
	if future.Error() != nil {
		return &proto.Empty{}, future.Error()
	}

	log.Println("[proto] (ConsensusRemove) request successful")
	return &proto.Empty{}, nil
}
