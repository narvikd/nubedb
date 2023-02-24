package fsm

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/narvikd/errorskit"
)

func (dbFSM DatabaseFSM) BackupDB() ([]byte, error) {
	w := bytes.NewBuffer(make([]byte, 0))
	_, err := dbFSM.db.Backup(w, 0)
	if err != nil {
		return nil, err
	}
	if w.Len() <= 0 {
		return nil, errors.New("backup size is 0 or lower")
	}
	return w.Bytes(), nil
}

func (dbFSM DatabaseFSM) RestoreDB(contents any) error {
	// maxPendingWrites is 256, to minimise memory usage and overall finish time. This is just a throttle option.
	const maxPendingWrites = 256

	dbValue, errMarshal := json.Marshal(contents)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal value on set")
	}

	return dbFSM.db.Load(bytes.NewReader(dbValue), maxPendingWrites)
}
