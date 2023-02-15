package protoserver

import (
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
	"net"
	"nubedb/api/proto"
	"nubedb/cluster/consensus/fsm"
	"nubedb/internal/app"
)

type server struct {
	proto.UnimplementedServiceServer
	Consensus *raft.Raft
	FSM       *fsm.DatabaseFSM
}

func Start(a *app.App) error {
	listen, errListen := net.Listen("tcp", a.Config.GrpcAddress)
	if errListen != nil {
		return errListen
	}
	srvModel := new(server)

	protoServer := grpc.NewServer()
	proto.RegisterServiceServer(protoServer, srvModel)
	srvModel.Consensus = a.Node.Consensus
	srvModel.FSM = a.Node.FSM

	errServe := protoServer.Serve(listen)
	if errServe != nil {
		return errServe
	}

	return nil
}
