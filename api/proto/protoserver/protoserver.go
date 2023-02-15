package protoserver

import (
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"net"
	"nubedb/api/proto"
	"nubedb/cluster/consensus/fsm"
	"nubedb/internal/app"
	"nubedb/internal/config"
)

type server struct {
	proto.UnimplementedServiceServer
	Config    config.Config
	Consensus *raft.Raft
	FSM       *fsm.DatabaseFSM
}

func Start(a *app.App) error {
	listen, errListen := net.Listen("tcp", a.Config.CurrentNode.GrpcAddress)
	if errListen != nil {
		return errListen
	}

	srvModel := &server{
		Config:    a.Config,
		Consensus: a.Node.Consensus,
		FSM:       a.Node.FSM,
	}

	protoServer := grpc.NewServer()
	proto.RegisterServiceServer(protoServer, srvModel)

	return protoServer.Serve(listen)
}
