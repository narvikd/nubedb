package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster/consensus/fsm"
	"nubedb/internal/config"
	"time"
)

const (
	errDBCluster = "consensus returned an error when trying to Apply an order."
)

func Execute(cfg config.Config, consensus *raft.Raft, payload *fsm.Payload) error {
	payloadData, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to the DB cluster")
	}

	if consensus.State() != raft.Leader {
		return handleForwardLeaderFuture(cfg, consensus, payload)
	}

	return ApplyLeaderFuture(consensus, payloadData)
}

func ApplyLeaderFuture(consensus *raft.Raft, payloadData []byte) error {
	const timeout = 500 * time.Millisecond

	if consensus.State() != raft.Leader {
		return errors.New("node is not a leader")
	}

	future := consensus.Apply(payloadData, timeout)
	if future.Error() != nil {
		return errorskit.Wrap(future.Error(), errDBCluster+" At future")
	}

	response := future.Response().(*fsm.ApplyRes)
	if response.Error != nil {
		return errorskit.Wrap(response.Error, errDBCluster+" At response")
	}

	return nil
}

func handleForwardLeaderFuture(cfg config.Config, consensus *raft.Raft, payload *fsm.Payload) error {
	_, leaderID := consensus.LeaderWithID()
	leaderCfg := cfg.Nodes[string(leaderID)]
	err := forwardLeaderFuture(leaderCfg, payload)
	if err != nil {
		return errorskit.Wrap(err, errDBCluster+" At handling forward")
	}
	return nil
}

func forwardLeaderFuture(leaderCfg config.NodeCfg, payload *fsm.Payload) error {
	log.Printf("[proto] payload for leader received in this node, forwarding to leader '%s' @ '%s'\n",
		leaderCfg.ID, leaderCfg.GrpcAddress,
	)

	payloadData, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to the Leader's DB cluster")
	}

	const errGrpcTalk = "failed to get an ok response from the Leader via grpc"

	conn, errDial := grpc.Dial(leaderCfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return errorskit.Wrap(errDial, "failed to connect to the Leader via grpc")
	}
	defer conn.Close()

	client := proto.NewServiceClient(conn)
	_, errTalk := client.ExecuteOnLeader(context.Background(), &proto.Request{
		Payload: payloadData,
	})
	if errTalk != nil {
		return errorskit.Wrap(errTalk, errGrpcTalk)
	}

	return nil
}
