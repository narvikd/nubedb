package protoserver

import (
	"context"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster"
)

func (srv *server) ExecuteOnLeader(ctx context.Context, req *proto.Request) (*proto.Empty, error) {
	log.Println("[proto] request received, processing...")

	errExecute := cluster.ApplyLeaderFuture(srv.Consensus, req.Payload)
	if errExecute != nil {
		return &proto.Empty{}, errExecute
	}

	log.Println("[proto] request successful")

	return &proto.Empty{}, nil
}
