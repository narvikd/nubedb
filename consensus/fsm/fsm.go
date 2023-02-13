package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"io"
)

type DatabaseFSM struct {
	db *badger.DB
}

type snapshot struct{}

// Payload is the Payload sent for use in raft.Apply
type Payload struct {
	Key       string `json:"key" validate:"required"`
	Value     any    `json:"value"`
	Operation string `json:"operation"`
}

// ApplyResponse represents the response from raft's apply
type ApplyResponse struct {
	Error error
	Data  interface{}
}

func New(db *badger.DB) raft.FSM {
	return &DatabaseFSM{db: db}
}

func (dbFSM DatabaseFSM) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		p := new(Payload)
		errUnMarshal := json.Unmarshal(log.Data, p)
		if errUnMarshal != nil {
			return errorskit.Wrap(errUnMarshal, "couldn't unmarshal storage payload")
		}

		switch p.Operation {
		case "SET":
			//dbFSM.db.Store(p.Key, p.Value)
			return &ApplyResponse{}
		default:
			return &ApplyResponse{
				Error: fmt.Errorf("operation type not recognized: %v", p.Operation),
			}
		}
	default:
		return fmt.Errorf("raft type not recognized: %v", log.Type)
	}
}

func (dbFSM DatabaseFSM) Restore(snap io.ReadCloser) error {
	d := json.NewDecoder(snap)
	for d.More() {
		mapper := map[string]any{}
		errDecode := d.Decode(&mapper)
		if errDecode != nil {
			return errorskit.Wrap(errDecode, "couldn't decode snapshot")
		}

		//for k, v := range mapper {
		//	dbFSM.db.Store(k, v)
		//}
	}

	return snap.Close()
}

// Snapshot isn't needed because BadgerDB persists data when Apply is called.
func (dbFSM DatabaseFSM) Snapshot() (raft.FSMSnapshot, error) {
	return snapshot{}, nil
}

// Persist isn't needed because BadgerDB persists data when Apply is called.
func (s snapshot) Persist(_ raft.SnapshotSink) error {
	return nil
}

func (s snapshot) Release() {}
