package fsm

import "github.com/narvikd/errorskit"

func (dbFSM DatabaseFSM) delete(k string) error {
	txn := dbFSM.db.NewTransaction(true)
	defer txn.Discard()
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
