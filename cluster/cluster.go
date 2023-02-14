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
	data, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to the DB cluster")
	}

	if consensus.State() == raft.Leader {
		return ApplyLeaderFuture(consensus, data)
	}

	return errors.New(errNodeNotALeader)
}

func ApplyLeaderFuture(consensus *raft.Raft, data []byte) error {
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
