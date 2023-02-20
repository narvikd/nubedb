// Package protoserver provides nubedb's gRPC server implementation.
package protoserver

import (
	"google.golang.org/grpc"
	"net"
	"nubedb/api/proto"
	"nubedb/cluster/consensus"
	"nubedb/internal/app"
	"nubedb/internal/config"
)

// server represents the gRPC server.
type server struct {
	proto.UnimplementedServiceServer
	Config config.Config
	Node   *consensus.Node
}

// Start starts the gRPC server.
func Start(a *app.App) error {
	listen, errListen := net.Listen("tcp", a.Config.CurrentNode.GrpcAddress)
	if errListen != nil {
		return errListen
	}

	// Create the server model.
	srvModel := &server{
		Config: a.Config,
		Node:   a.Node,
	}

	protoServer := grpc.NewServer()
	proto.RegisterServiceServer(protoServer, srvModel) // register the server model

	return protoServer.Serve(listen)
}
