package protoserver

import (
	"google.golang.org/grpc"
	"net"
	"nubedb/api/proto"
	"nubedb/cluster/consensus"
	"nubedb/internal/app"
	"nubedb/internal/config"
)

type server struct {
	proto.UnimplementedServiceServer
	Config config.Config
	Node   *consensus.Node
}

func Start(a *app.App) error {
	listen, errListen := net.Listen("tcp", a.Config.CurrentNode.GrpcAddress)
	if errListen != nil {
		return errListen
	}

	srvModel := &server{
		Config: a.Config,
		Node:   a.Node,
	}

	protoServer := grpc.NewServer()
	proto.RegisterServiceServer(protoServer, srvModel)

	return protoServer.Serve(listen)
}
