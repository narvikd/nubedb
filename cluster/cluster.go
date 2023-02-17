package cluster

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"log"
	"nubedb/api/proto"
	"nubedb/api/proto/protoclient"
	"nubedb/cluster/consensus/fsm"
	"nubedb/internal/config"
	"time"
)

const (
	errDBCluster      = "consensus returned an error when trying to Apply an order."
	errGrpcTalkLeader = "failed to get an ok response from the Leader via grpc"
	errGrpcTalkNode   = "failed to get an ok response from the node via grpc"
)

func Execute(consensus *raft.Raft, payload *fsm.Payload) error {
	payloadData, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to the DB cluster")
	}

	if consensus.State() != raft.Leader {
		return handleForwardLeaderFuture(consensus, payload)
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

func handleForwardLeaderFuture(consensus *raft.Raft, payload *fsm.Payload) error {
	_, leaderID := consensus.LeaderWithID()
	err := forwardLeaderFuture(string(leaderID), payload)
	if err != nil {
		return errorskit.Wrap(err, errDBCluster+" At handling forward")
	}
	return nil
}

func forwardLeaderFuture(leaderID string, payload *fsm.Payload) error {
	leaderGrpcAddr := config.MakeGrpcAddress(leaderID)
	log.Printf("[proto] payload for leader received in this node, forwarding to leader '%s' @ '%s'\n",
		leaderID, leaderGrpcAddr,
	)

	payloadData, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to the Leader's DB cluster")
	}

	conn, errConn := protoclient.NewConnection(leaderGrpcAddr)
	if errConn != nil {
		return errConn
	}
	defer conn.Cleanup()

	_, errTalk := conn.Client.ExecuteOnLeader(conn.Ctx, &proto.ExecuteOnLeaderRequest{
		Payload: payloadData,
	})
	if errTalk != nil {
		return errorskit.Wrap(errTalk, errGrpcTalkLeader)
	}

	return nil
}

func IsLeader(addr string) (bool, error) {
	conn, errConn := protoclient.NewConnection(addr)
	if errConn != nil {
		return false, errConn
	}
	defer conn.Cleanup()

	res, errTalk := conn.Client.IsLeader(conn.Ctx, &proto.Empty{})
	if errTalk != nil {
		return false, errorskit.Wrap(errTalk, errGrpcTalkNode)
	}

	return res.IsLeader, nil
}

func ConsensusJoin(nodeID string, nodeConsensusAddr string, leaderGrpcAddr string) error {
	// TODO: Maybe add a message to know one node is contacting the other for this operation

	conn, errConn := protoclient.NewConnection(leaderGrpcAddr)
	if errConn != nil {
		return errConn
	}
	defer conn.Cleanup()

	_, errTalk := conn.Client.ConsensusJoin(conn.Ctx, &proto.ConsensusRequest{
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

	conn, errConn := protoclient.NewConnection(leaderGrpcAddr)
	if errConn != nil {
		return errConn
	}
	defer conn.Cleanup()

	_, errTalk := conn.Client.ConsensusRemove(conn.Ctx, &proto.ConsensusRequest{
		NodeID: nodeID,
	})
	if errTalk != nil {
		return errorskit.Wrap(errTalk, errGrpcTalkLeader)
	}

	return nil
}
