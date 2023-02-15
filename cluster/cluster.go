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
	timeoutGrpcCall      = 3 * time.Second
	errDBCluster         = "consensus returned an error when trying to Apply an order."
	errGrpcTalkLeader    = "failed to get an ok response from the Leader via grpc"
	errGrpcConnectLeader = "failed to connect to the Leader via grpc"
	errGrpcTalkNode      = "failed to get an ok response from the node via grpc"
	errGrpcConnectNode   = "failed to connect to the node via grpc"
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

	// Hardcoded since it's just for dial
	ctxDial, cancelCtxDial := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxDial()

	conn, errDial := grpc.DialContext(ctxDial, leaderCfg.GrpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return errorskit.Wrap(errDial, errGrpcConnectLeader)
	}
	defer conn.Close()

	ctxExecuteCall, cancelCtxExecuteCall := context.WithTimeout(context.Background(), timeoutGrpcCall)
	defer cancelCtxExecuteCall()

	_, errTalk := proto.NewServiceClient(conn).ExecuteOnLeader(ctxExecuteCall, &proto.ExecuteOnLeaderRequest{
		Payload: payloadData,
	})
	if errTalk != nil {
		return errorskit.Wrap(errTalk, errGrpcTalkLeader)
	}

	return nil
}

func IsLeader(addr string) (bool, error) {
	// Hardcoded since it's just for dial
	ctxDial, cancelCtxDial := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxDial()

	conn, errDial := grpc.DialContext(ctxDial, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return false, errorskit.Wrap(errDial, errGrpcConnectNode)
	}
	defer conn.Close()

	ctxExecuteCall, cancelCtxExecuteCall := context.WithTimeout(context.Background(), timeoutGrpcCall)
	defer cancelCtxExecuteCall()

	res, errTalk := proto.NewServiceClient(conn).IsLeader(ctxExecuteCall, &proto.Empty{})
	if errTalk != nil {
		return false, errorskit.Wrap(errTalk, errGrpcTalkNode)
	}

	return res.IsLeader, nil
}

func ConsensusJoin(nodeID string, nodeConsensusAddr string, leaderGrpcAddr string) error {
	// TODO: Maybe add a message to know one node is contacting the other for this operation

	// Hardcoded since it's just for dial
	ctxDial, cancelCtxDial := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxDial()

	conn, errDial := grpc.DialContext(ctxDial, leaderGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return errorskit.Wrap(errDial, errGrpcConnectLeader)
	}
	defer conn.Close()

	ctxExecuteCall, cancelCtxExecuteCall := context.WithTimeout(context.Background(), timeoutGrpcCall)
	defer cancelCtxExecuteCall()

	_, errTalk := proto.NewServiceClient(conn).ConsensusJoin(ctxExecuteCall, &proto.ConsensusRequest{
		NodeID:            nodeID,
		NodeConsensusAddr: nodeConsensusAddr,
	})
	if errTalk != nil {
		return errorskit.Wrap(errTalk, errGrpcTalkLeader)
	}

	return nil
}

func ConsensusRemove(nodeID string, leaderGrpcAddr string) error {
	// TODO: Maybe add a message to know one node is contacting the other for this operation

	// Hardcoded since it's just for dial
	ctxDial, cancelCtxDial := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelCtxDial()

	conn, errDial := grpc.DialContext(ctxDial, leaderGrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return errorskit.Wrap(errDial, errGrpcConnectLeader)
	}
	defer conn.Close()

	ctxExecuteCall, cancelCtxExecuteCall := context.WithTimeout(context.Background(), timeoutGrpcCall)
	defer cancelCtxExecuteCall()

	_, errTalk := proto.NewServiceClient(conn).ConsensusRemove(ctxExecuteCall, &proto.ConsensusRequest{
		NodeID: nodeID,
	})
	if errTalk != nil {
		return errorskit.Wrap(errTalk, errGrpcTalkLeader)
	}

	return nil
}
