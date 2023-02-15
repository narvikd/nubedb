package protoserver

import (
	"context"
	"encoding/json"
	"github.com/narvikd/errorskit"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster"
	"nubedb/cluster/consensus/fsm"
)

func (srv *server) ExecuteOnLeader(ctx context.Context, req *proto.Request) (*proto.Empty, error) {
	log.Println("[proto] request received, processing...")

	p := new(fsm.Payload)
	errUnmarshal := json.Unmarshal(req.Payload, p)
	if errUnmarshal != nil {
		return &proto.Empty{}, errorskit.Wrap(errUnmarshal, "couldn't unmarshal payload")
	}

	errExecute := cluster.Execute(srv.Config, srv.Consensus, p)
	if errExecute != nil {
		return &proto.Empty{}, errExecute
	}

	log.Println("[proto] request successful")

	return &proto.Empty{}, nil
}
