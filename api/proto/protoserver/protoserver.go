package protoserver

import (
	"fmt"
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
	listen, errListen := net.Listen("tcp", makeAddr(a.Config.Host, a.Config.GrpcPort))
	if errListen != nil {
		return errListen
	}
	srvModel := new(server)

	protoServer := grpc.NewServer()
	proto.RegisterServiceServer(protoServer, srvModel)

	errServe := protoServer.Serve(listen)
	if errServe != nil {
		return errServe
	}

	srvModel.Consensus = a.Node.Consensus
	srvModel.FSM = a.Node.FSM

	return nil
}

func makeAddr(host string, port int) string {
	if host == "0.0.0.0" {
		host = "" // #nosec G102 -- This is needed for listening in docker
	}
	return fmt.Sprintf("%s:%v", host, port)
}
