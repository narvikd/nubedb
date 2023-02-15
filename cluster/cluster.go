package cluster

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"nubedb/cluster/consensus/fsm"
	"time"
)

const errNodeNotALeader = "node is not a leader"

func Execute(consensus *raft.Raft, payload *fsm.Payload) error {
	payloadData, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to the DB cluster")
	}

	if consensus.State() != raft.Leader {
		return errors.New(errNodeNotALeader)
	}

	return applyLeaderFuture(consensus, payloadData)
}

func applyLeaderFuture(consensus *raft.Raft, payloadData []byte) error {
	const timeout = 500 * time.Millisecond

	future := consensus.Apply(payloadData, timeout)
	if future.Error() != nil {
		return errorskit.Wrap(future.Error(), "couldn't persist data to DB Cluster")
	}

	response := future.Response().(*fsm.ApplyRes)
	if response.Error != nil {
		return errorskit.Wrap(response.Error, "DB cluster returned an error when trying to persist data to it")
	}

	return nil
}

//func forwardsLeaderFuture(consensus *raft.Raft, payloadData []byte) error {
//	// Regex for port: /(?<=:)\d+/
//
//	//log.Printf("[proto] payload for leader received in this node, forwarding to leader '%s' @ '%s'\n",
//	//	leaderID, leaderAddress,
//	//)
//
//	const errGrpcTalk = "failed to get an ok response from the Leader via grpc"
//
//	conn, errDial := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if errDial != nil {
//		return errorskit.Wrap(errDial, "failed to connect to the Leader via grpc")
//	}
//	defer conn.Close()
//
//	client := proto.NewServiceClient(conn)
//	res, errTalk := client.ExecuteOnLeader(context.Background(), &proto.Request{
//		Payload: payloadData,
//	})
//	if errTalk != nil {
//		return errorskit.Wrap(errTalk, errGrpcTalk)
//	}
//	if res.Error != "" {
//		return errors.New(errGrpcTalk)
//	}
//
//	return nil
//}
