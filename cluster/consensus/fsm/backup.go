package fsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/narvikd/errorskit"
)

func (dbFSM DatabaseFSM) BackupDB() ([]byte, error) {
	m := make(map[string]any)
	txn := dbFSM.db.NewTransaction(false)
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()

		// Get the key and value
		key := item.KeyCopy(nil)
		var value []byte
		errVal := item.Value(func(val []byte) error {
			value = append([]byte{}, val...)
			return nil
		})
		if errVal != nil {
			return nil, errVal
		}

		// json.RawMessage prevents "any" types to be converted to string
		m[string(key)] = json.RawMessage(value)
	}

	b, errJson := json.Marshal(m)
	if errJson != nil {
		return nil, errorskit.Wrap(errJson, "couldn't marshal backup")
	}

	return b, nil
}

func (dbFSM DatabaseFSM) RestoreDB(contents any) error {
	m := contents.(map[string]any)
	txn := dbFSM.db.NewTransaction(true)
	defer txn.Discard()

	for k, v := range m {
		dbValue, errMarshalValue := json.Marshal(v)
		if errMarshalValue != nil {
			errMsg := fmt.Sprintf("couldn't restore key '%s'. Err: %v", k, errMarshalValue)
			return errors.New(errMsg)
		}

		errSet := txn.Set([]byte(k), dbValue)
		if errSet != nil {
			return errorskit.Wrap(errSet, "couldn't set on restore")
		}

	}

	errCommit := txn.Commit()
	if errCommit != nil {
		return errorskit.Wrap(errCommit, "couldn't commit transaction")
	}

	return nil
}
