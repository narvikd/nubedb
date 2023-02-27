package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"io"
	"time"
)

// DatabaseFSM represents the finite state machine implementation for the database
type DatabaseFSM struct {
	db *badger.DB
}

// snapshot's is a struct that represents the snapshot of the state machine.
//
// In nubedb's particular case, since it uses BadgerDB, it already persists data when Apply is called,
//
// the Snapshot method is not needed.
type snapshot struct{}

// Payload is the Payload sent for use in raft.Apply
type Payload struct {
	Key       string `json:"key" validate:"required"`
	Value     any    `json:"value"`
	Operation string `json:"operation"`
}

// ApplyRes represents the response from raft.Apply
type ApplyRes struct {
	Data  any
	Error error
}

// Apply processes a Raft log entry
func (dbFSM DatabaseFSM) Apply(log *raft.Log) any {
	// Process the Raft log entry based on its type
	switch log.Type {
	case raft.LogCommand:
		p := new(Payload)
		errUnMarshal := json.Unmarshal(log.Data, p)
		if errUnMarshal != nil {
			return errorskit.Wrap(errUnMarshal, "couldn't unmarshal storage payload")
		}

		// Process the log entry based on the operation type
		// &ApplyRes struct is used to represent the response from the Apply method of the Raft log
		switch p.Operation {
		case "SET":
			return &ApplyRes{
				Error: dbFSM.set(p.Key, p.Value),
			}
		case "DELETE":
			return &ApplyRes{
				Error: dbFSM.delete(p.Key),
			}
		case "RESTOREDB":
			return &ApplyRes{
				Error: dbFSM.RestoreDB(p.Value),
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

// Restore restores the finite state machine from a snapshot.
//
// io.ReadCloser represents a snapshot of the state machine that needs to be restored.
func (dbFSM DatabaseFSM) Restore(snap io.ReadCloser) error {
	d := json.NewDecoder(snap)
	// Loops through the snapshot data and decodes each key-value pair into a Payload struct.
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

// Snapshot is part of the raft.FSM interface, and it's used to create a snapshot of the current state of the system.
//
// This is used later to restore the state of the system.
//
// In the case of nubedb, this method isn't needed because BadgerDB persists data when Apply is called.
//
// Therefore, the Snapshot method simply returns an empty snapshot and a nil error.
//
// Note: A snapshot can be useful when the log becomes too large, and it is necessary to reduce its size.
//
// When a snapshot is created, it includes all the data required to restore the system to its current state,
// allowing the system to discard all the log entries prior to the snapshot.
func (dbFSM DatabaseFSM) Snapshot() (raft.FSMSnapshot, error) {
	return snapshot{}, nil
}

// Persist method is called when a new snapshot is taken, it provides a way to save the snapshot somewhere (for example to a disk).
//
// In the case of the nubedb, Persist does nothing because BadgerDB already persists data when Apply is called.
func (s snapshot) Persist(_ raft.SnapshotSink) error {
	return nil
}

// Release method is used to release any resources that were acquired by the Persist method.
//
// In the case of nubedb, the Release method is empty because Persist does not acquire any resources.
func (s snapshot) Release() {}

// New creates a new instance of DatabaseFSM.
//
// Check DatabaseFSM for more info
func New(storageDir string) (*DatabaseFSM, error) {
	database, err := newDB(storageDir)
	if err != nil {
		return nil, err
	}
	return &DatabaseFSM{db: database}, nil
}

func newDB(storageDir string) (*badger.DB, error) {
	db, err := badger.Open(badger.DefaultOptions(storageDir))
	if err != nil {
		return nil, errorskit.Wrap(err, "couldn't open badgerDB")
	}
	go badgerGC(db)
	return db, nil
}

func badgerGC(db *badger.DB) {
	const (
		gcCycle      = 15 * time.Minute
		discardRatio = 0.5
	)
	ticker := time.NewTicker(gcCycle)
	defer ticker.Stop()
	for range ticker.C {
	again:
		err := db.RunValueLogGC(discardRatio)
		if err == nil {
			goto again
		}
	}
}
