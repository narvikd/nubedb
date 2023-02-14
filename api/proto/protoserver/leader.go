package protoserver

import (
	"context"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster"
)

func (srv *server) LeaderApplyFuture(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	log.Println("[proto] request received")
	err := cluster.ApplyLeaderFuture(srv.Consensus, req.Payload)
	if err != nil {
		return &proto.Response{
			Error: err.Error(),
		}, err
	}

	log.Println("[proto] request successful")

	return &proto.Response{
		Error: "",
	}, nil
}
