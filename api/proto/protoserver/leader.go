package protoserver

import (
	"context"
	"github.com/hashicorp/raft"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster"
)

func (srv *server) ExecuteOnLeader(ctx context.Context, req *proto.ExecuteOnLeaderRequest) (*proto.Empty, error) {
	log.Println("[proto](ExecuteOnLeader) request received, processing...")

	errExecute := cluster.ApplyLeaderFuture(srv.Consensus, req.Payload)
	if errExecute != nil {
		return &proto.Empty{}, errExecute
	}

	log.Println("[proto](ExecuteOnLeader) request successful")

	return &proto.Empty{}, nil
}

func (srv *server) IsLeader(ctx context.Context, req *proto.Empty) (*proto.IsLeaderResponse, error) {
	log.Println("[proto] (IsLeader) request received, processing...")
	is := srv.Consensus.State() == raft.Leader
	log.Println("[proto](ExecuteOnLeader) request successful")
	return &proto.IsLeaderResponse{IsLeader: is}, nil
}
