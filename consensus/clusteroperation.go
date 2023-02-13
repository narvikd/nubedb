package consensus

import (
	"encoding/json"
	"errors"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"nubedb/consensus/fsm"
	"time"
)

func ClusterOperation(consensus *raft.Raft, payload *fsm.Payload, operationType string) error {
	const timeout = 500 * time.Millisecond
	if consensus.State() != raft.Leader {
		return errors.New("node is not a leader")
	}

	payload.Operation = operationType

	data, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal data to send it to DB cluster")
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
