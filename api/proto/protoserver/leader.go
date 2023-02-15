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

func (srv *server) ExecuteOnLeader(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	log.Println("[proto] request received")

	p := new(fsm.Payload)
	errUnmarshal := json.Unmarshal(req.Payload, p)
	if errUnmarshal != nil {
		return &proto.Response{
			Error: errorskit.Wrap(errUnmarshal, "couldn't unmarshal payload").Error(),
		}, errUnmarshal
	}

	errExecute := cluster.Execute(srv.Consensus, p)
	if errExecute != nil {
		return &proto.Response{
			Error: errExecute.Error(),
		}, errExecute
	}

	log.Println("[proto] request successful")

	return &proto.Response{
		Error: "",
	}, nil
}
