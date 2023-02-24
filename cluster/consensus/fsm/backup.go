package fsm

import (
	"bytes"
	"errors"
)

func (dbFSM DatabaseFSM) Backup() ([]byte, error) {
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
