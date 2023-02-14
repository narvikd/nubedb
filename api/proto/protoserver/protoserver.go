package protoserver

import (
	"fmt"
	"google.golang.org/grpc"
	"net"
	"nubedb/api/proto"
	"nubedb/internal/config"
)

type server struct {
	proto.UnimplementedServiceServer
}

func Start(cfg config.Config) error {
	listen, errListen := net.Listen("tcp", makeAddr(cfg.Host, cfg.GrpcPort))
	if errListen != nil {
		return errListen
	}
	srv := grpc.NewServer()
	proto.RegisterServiceServer(srv, &server{})
	return srv.Serve(listen)
}

func makeAddr(host string, port int) string {
	if host == "0.0.0.0" {
		host = "" // #nosec G102 -- This is needed for listening in docker
	}
	return fmt.Sprintf("%s:%v", host, port)
}
