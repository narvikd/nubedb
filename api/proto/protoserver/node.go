package protoserver

import (
	"context"
	"log"
	"nubedb/api/proto"
)

func (srv *server) ReinstallNode(ctx context.Context, req *proto.Empty) (*proto.Empty, error) {
	log.Println("[proto] (Reset Node) request received, processing...")
	go srv.Node.ReinstallNode()
	log.Println("[proto] (Reset Node) request successful")
	return &proto.Empty{}, nil
}
