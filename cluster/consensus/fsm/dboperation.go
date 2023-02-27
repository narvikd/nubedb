package fsm

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/narvikd/errorskit"
)

// Get is a DatabaseFSM's method which gets a value from a key from the LOCAL NODE.
//
// This method isn't committed since there's no need for it.
func (dbFSM DatabaseFSM) Get(k string) (any, error) {
	var result any
	dbResultValue := make([]byte, 0)

	txn := dbFSM.db.NewTransaction(false)
	defer txn.Discard()
	dbResult, errGet := txn.Get([]byte(k))
	if errGet != nil {
		return nil, errGet
	}

	errDBResultValue := dbResult.Value(func(val []byte) error {
		dbResultValue = append(dbResultValue, val...)
		return nil
	})
	if errDBResultValue != nil {
		return nil, errDBResultValue
	}

	if dbResultValue == nil || len(dbResultValue) <= 0 {
		return nil, errors.New("no result for key")
	}

	errUnmarshal := json.Unmarshal(dbResultValue, &result)
	if errUnmarshal != nil {
		return nil, errorskit.Wrap(errUnmarshal, "couldn't unmarshal get results from DB")
	}

	errCommit := txn.Commit()
	if errCommit != nil {
		return nil, errorskit.Wrap(errCommit, "couldn't commit transaction")
	}

	return result, nil
}

func (dbFSM DatabaseFSM) GetKeys() []string {
	var keys []string
	txn := dbFSM.db.NewTransaction(false)
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	for it.Rewind(); it.Valid(); it.Next() {
		key := it.Item().KeyCopy(nil)
		keys = append(keys, string(key))
	}
	return keys
}

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
