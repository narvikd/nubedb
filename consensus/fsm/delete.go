package fsm

import (
	"errors"
	"github.com/narvikd/errorskit"
)

func (dbFSM DatabaseFSM) delete(k string) error {
	txn := dbFSM.db.NewTransaction(true)
	defer txn.Discard()

	dbResultValue := make([]byte, 0)
	dbResult, errGet := txn.Get([]byte(k))
	if errGet != nil {
		return errGet
	}

	errDBResultValue := dbResult.Value(func(val []byte) error {
		dbResultValue = append(dbResultValue, val...)
		return nil
	})

	if errDBResultValue != nil {
		return errDBResultValue
	}

	if dbResultValue == nil || len(dbResultValue) <= 0 {
		return errors.New("no result for key")
	}

	err := txn.Delete([]byte(k))
	if err != nil {
		return err
	}

	errCommit := txn.Commit()
	if errCommit != nil {
		return errorskit.Wrap(errCommit, "couldn't commit transaction")
	}

	return nil
}
