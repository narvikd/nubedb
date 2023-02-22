package fsm

import (
	"github.com/narvikd/errorskit"
)

// delete is a DatabaseFSM's method which deletes a key-value pair from the database.
func (dbFSM DatabaseFSM) delete(k string) error {
	txn := dbFSM.db.NewTransaction(true)
	defer txn.Discard()

	// Get the value for the key to check if it exists (it will return an error if it doesn't)
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
