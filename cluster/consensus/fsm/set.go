package fsm

import (
	"encoding/json"
	"errors"
	"github.com/narvikd/errorskit"
)

// set is a DatabaseFSM's method which adds a key-value pair to the database.
func (dbFSM DatabaseFSM) set(k string, value any) error {
	dbValue, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		return errorskit.Wrap(errMarshal, "couldn't marshal value on set")
	}

	if dbValue == nil || len(dbValue) <= 0 {
		return errors.New("value was empty")
	}

	txn := dbFSM.db.NewTransaction(true)
	defer txn.Discard()
	errSet := txn.Set([]byte(k), dbValue)
	if errSet != nil {
		return errSet
	}

	errCommit := txn.Commit()
	if errCommit != nil {
		return errorskit.Wrap(errCommit, "couldn't commit transaction")
	}

	return nil
}
