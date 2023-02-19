package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"io"
	"nubedb/pkg/errorskit"
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

// ApplyRes represents the response from raft's apply
type ApplyRes struct {
	Data  any
	Error error
}

func New(db *badger.DB) *DatabaseFSM {
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
			return &ApplyRes{
				Error: dbFSM.set(p.Key, p.Value),
			}
		case "DELETE":
			return &ApplyRes{
				Error: dbFSM.delete(p.Key),
			}
		default:
			return &ApplyRes{
				Error: fmt.Errorf("operation type not recognized: %v", p.Operation),
			}
		}
	default:
		return fmt.Errorf("raft command not recognized: %v", log.Type)
	}
}

func (dbFSM DatabaseFSM) Restore(snap io.ReadCloser) error {
	d := json.NewDecoder(snap)
	for d.More() {
		dbValue := new(Payload)
		errDecode := d.Decode(&dbValue)
		if errDecode != nil {
			return errorskit.Wrap(errDecode, "couldn't decode snapshot")
		}

		errSet := dbFSM.set(dbValue.Key, dbValue.Value)
		if errSet != nil {
			return errorskit.Wrap(errSet, "couldn't restore key while restoring a snapshot")
		}
	}

	// Once the loop has completed, the program has to call the Token method on the decoder to check
	// if it finds the closing bracket from the data stream.
	// This particular token should not return an error, meaning that the program has reached the end of the data stream
	// and all other tokens have been extracted.
	// If it indeed returns an error, it would mean that an unexpected token was encountered during the Snapshot
	// restoration, such as a closing brace in the wrong place.
	// It could also be an indicator that the end of the input stream was reached before the closing bracket was found.
	_, errToken := d.Token()
	if errToken != nil && errToken != io.EOF { // If we reach the end of the stream, it isn't an error
		return errorskit.Wrap(errToken, "couldn't restore snapshot due to json malformation")
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
