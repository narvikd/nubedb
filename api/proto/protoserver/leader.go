package protoserver

import (
	"context"
	"errors"
	"github.com/hashicorp/raft"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster"
)

func (srv *server) ExecuteOnLeader(ctx context.Context, req *proto.ExecuteOnLeaderRequest) (*proto.Empty, error) {
	log.Println("[proto] (ExecuteOnLeader) request received, processing...")

	errExecute := cluster.ApplyLeaderFuture(srv.Consensus, req.Payload)
	if errExecute != nil {
		return &proto.Empty{}, errExecute
	}

	log.Println("[proto] (ExecuteOnLeader) request successful")
	return &proto.Empty{}, nil
}

func (srv *server) IsLeader(ctx context.Context, req *proto.Empty) (*proto.IsLeaderResponse, error) {
	log.Println("[proto] (IsLeader) request received, processing...")
	is := srv.Consensus.State() == raft.Leader
	log.Println("[proto] (IsLeader) request successful")
	return &proto.IsLeaderResponse{IsLeader: is}, nil
}

func (srv *server) ConsensusJoin(ctx context.Context, req *proto.ConsensusRequest) (*proto.Empty, error) {
	log.Println("[proto] (ConsensusJoin) request received, processing...")

	consensusCfg := srv.Consensus.GetConfiguration().Configuration()
	for _, s := range consensusCfg.Servers {
		if req.NodeID == string(s.ID) {
			return &proto.Empty{}, errors.New("node was already part of the network")
		}
	}

	future := srv.Consensus.AddVoter(raft.ServerID(req.NodeID), raft.ServerAddress(req.NodeConsensusAddr), 0, 0)
	if future.Error() != nil {
		return &proto.Empty{}, future.Error()
	}

	log.Println("[proto] (ConsensusJoin) request successful")
	return &proto.Empty{}, nil
}

func (srv *server) ConsensusRemove(ctx context.Context, req *proto.ConsensusRequest) (*proto.Empty, error) {
	log.Println("[proto] (ConsensusRemove) request received, processing...")

	future := srv.Consensus.RemoveServer(raft.ServerID(req.NodeID), 0, 0)
	if future.Error() != nil {
		return &proto.Empty{}, future.Error()
	}

	log.Println("[proto] (ConsensusRemove) request successful")
	return &proto.Empty{}, nil
}
