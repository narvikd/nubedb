package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"nubedb/api/proto"
	"nubedb/cluster/consensus/fsm"
	"time"
)

const errNodeNotALeader = "node is not a leader"

func Execute(consensus *raft.Raft, payload *fsm.Payload) error {
	data, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to the DB cluster")
	}

	if consensus.State() == raft.Leader {
		return applyLeaderFuture(consensus, data)
	}

	return forwardToRemoteLeader(consensus, data)
}

func applyLeaderFuture(consensus *raft.Raft, data []byte) error {
	const timeout = 500 * time.Millisecond

	if consensus.State() != raft.Leader {
		return errors.New(errNodeNotALeader)
	}

	future := consensus.Apply(data, timeout)
	if future.Error() != nil {
		return errorskit.Wrap(future.Error(), "couldn't persist data to DB Cluster")
	}

	response := future.Response().(*fsm.ApplyRes)
	if response.Error != nil {
		return errorskit.Wrap(response.Error, "DB cluster returned an error when trying to persist data to it")
	}

	return nil
}

func forwardToRemoteLeader(consensus *raft.Raft, payloadData []byte) error {
	_, leaderID := consensus.LeaderWithID()
	fmt.Println(leaderID)
	leaderAddress := ""

	switch leaderID {
	case "node1":
		leaderAddress = "127.0.0.1:5001"
		break
	case "node2":
		leaderAddress = "127.0.0.1:5002"
		break
	case "node3":
		leaderAddress = "127.0.0.1:5003"
		break
	}

	if leaderAddress == "" || leaderID == "" {
		return errors.New("tried to contact remote leader but couldn't find his address")
	}

	log.Printf("[proto] payload for leader received in this node, forwarding to leader '%s' @ '%s'\n",
		leaderID, leaderAddress,
	)

	const maxRetryCount = 3
	currentTry := 0
	currentSleep := 1 * time.Second

	// Retry pattern implementation
	for {
		errTalk := talkToRemoteLeader(leaderAddress, payloadData)
		if errTalk != nil {
			currentTry++
			if currentTry == 1 {
				log.Println("[proto] An error has occurred connecting to the Leader server. Retrying...")
			}

			if currentTry >= maxRetryCount {
				log.Printf(
					"[proto] Tried to contact the Leader server %v times and it didn't answered back. Aborting...\n",
					maxRetryCount,
				)
				return errTalk
			}

			log.Printf("[proto] request delayed for %v\n", currentSleep)
			time.Sleep(currentSleep)
			currentSleep *= 3
			continue
		}

		break
	}
	return nil
}

func talkToRemoteLeader(addr string, payloadData []byte) error {
	const errGrpcTalk = "failed to get an ok response from the Leader via grpc"

	conn, errDial := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if errDial != nil {
		return errorskit.Wrap(errDial, "failed to connect to the Leader via grpc")
	}
	defer conn.Close()

	client := proto.NewServiceClient(conn)
	res, errTalk := client.ExecuteOnLeader(context.Background(), &proto.Request{
		Payload: payloadData,
	})
	if errTalk != nil {
		return errorskit.Wrap(errTalk, errGrpcTalk)
	}
	if res.Error != "" {
		return errors.New(errGrpcTalk)
	}

	return nil
}
