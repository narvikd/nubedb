package fsm

import (
	"encoding/json"
	"github.com/narvikd/errorskit"
)

func (dbFSM DatabaseFSM) Get(k string) (map[string]any, bool, error) {
	resultMap := map[string]any{}
	dbResultValue := make([]byte, 0)

	txn := dbFSM.db.NewTransaction(false)
	dbResult, errGet := txn.Get([]byte(k))
	if errGet != nil {
		return nil, false, errGet
	}

	errDBResultValue := dbResult.Value(func(val []byte) error {
		dbResultValue = append(dbResultValue, val...)
		return nil
	})
	if errDBResultValue != nil {
		return nil, false, errDBResultValue
	}

	if dbResultValue == nil || len(dbResultValue) <= 0 {
		return nil, false, nil
	}

	errUnmarshal := json.Unmarshal(dbResultValue, &resultMap)
	if errUnmarshal != nil {
		return nil, false, errorskit.Wrap(errUnmarshal, "couldn't unmarshal get results from DB")
	}

	errCommit := txn.Commit()
	if errCommit != nil {
		return nil, false, errorskit.Wrap(errCommit, "couldn't commit transaction")
	}

	return resultMap, true, nil
}
