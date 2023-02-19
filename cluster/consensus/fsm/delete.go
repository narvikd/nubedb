package fsm

import (
	"nubedb/pkg/errorskit"
)

func (dbFSM DatabaseFSM) delete(k string) error {
	txn := dbFSM.db.NewTransaction(true)
	defer txn.Discard()

	_, errGet := txn.Get([]byte(k))
	if errGet != nil {
		return errGet
	}

	errDelete := txn.Delete([]byte(k))
	if errDelete != nil {
		return errDelete
	}

	errCommit := txn.Commit()
	if errCommit != nil {
		return errorskit.Wrap(errCommit, "couldn't commit transaction")
	}

	return nil
}
