package fsm

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"io"
	"sync"
)

type DatabaseFSM struct {
	DB *sync.Map
}

type snapshot struct {
	m *sync.Map
}

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

func New(db *sync.Map) *DatabaseFSM {
	return &DatabaseFSM{DB: db}
}

func (dbFSM *DatabaseFSM) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		p := new(Payload)
		errUnMarshal := json.Unmarshal(log.Data, p)
		if errUnMarshal != nil {
			return errorskit.Wrap(errUnMarshal, "couldn't unmarshal storage payload")
		}

		switch p.Operation {
		case "SET":
			dbFSM.DB.Store(p.Key, p.Value)
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

func (dbFSM *DatabaseFSM) Snapshot() (raft.FSMSnapshot, error) {
	return snapshot{dbFSM.DB}, nil
}

func (dbFSM *DatabaseFSM) Restore(snap io.ReadCloser) error {
	d := json.NewDecoder(snap)
	for d.More() {
		mapper := map[string]any{}
		errDecode := d.Decode(&mapper)
		if errDecode != nil {
			return errorskit.Wrap(errDecode, "couldn't decode snapshot")
		}

		for k, v := range mapper {
			dbFSM.DB.Store(k, v)
		}
	}

	return snap.Close()
}

func (s snapshot) Persist(sink raft.SnapshotSink) error {
	mapper := map[string]any{}
	s.m.Range(func(k, v any) bool {
		mapper[k.(string)] = v
		return true
	})

	defer sink.Close()

	err := json.NewEncoder(sink).Encode(mapper)
	if err != nil {
		_ = sink.Cancel()
		return errorskit.Wrap(err, "couldn't encode snapshot into map")
	}

	return nil
}

func (s snapshot) Release() {}
