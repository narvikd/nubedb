package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/narvikd/fiberparser"
	"io"
	"nubedb/api/rest/jsonresponse"
	"nubedb/cluster"
	"nubedb/cluster/consensus/fsm"
	"strings"
)

func (a *ApiCtx) storeGet(fiberCtx *fiber.Ctx) error {
	payload := new(fsm.Payload)
	errParse := fiberparser.ParseAndValidate(fiberCtx, payload)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}

	value, errGet := a.Node.FSM.Get(payload.Key)
	if errGet != nil {
		if strings.Contains(strings.ToLower(errGet.Error()), "key not found") {
			return jsonresponse.NotFound(fiberCtx, "key doesn't exist")
		}
		return jsonresponse.ServerError(fiberCtx, "couldn't get key from DB: "+errGet.Error())
	}

	return jsonresponse.OK(fiberCtx, "data retrieved successfully", value)
}

func (a *ApiCtx) storeSet(fiberCtx *fiber.Ctx) error {
	const operationType = "SET"

	payload := new(fsm.Payload)
	errParse := fiberparser.ParseAndValidate(fiberCtx, payload)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}
	payload.Operation = operationType

	errCluster := cluster.Execute(a.Node.Consensus, payload)
	if errCluster != nil {
		return jsonresponse.ServerError(fiberCtx, errCluster.Error())
	}

	return jsonresponse.OK(fiberCtx, "data persisted successfully", "")
}

func (a *ApiCtx) storeDelete(fiberCtx *fiber.Ctx) error {
	const operationType = "DELETE"

	payload := new(fsm.Payload)
	errParse := fiberparser.ParseAndValidate(fiberCtx, payload)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}
	payload.Operation = operationType

	errCluster := cluster.Execute(a.Node.Consensus, payload)
	if errCluster != nil {
		if strings.Contains(strings.ToLower(errCluster.Error()), "key not found") {
			return jsonresponse.NotFound(fiberCtx, "key doesn't exist")
		}
		return jsonresponse.ServerError(fiberCtx, errCluster.Error())
	}

	return jsonresponse.OK(fiberCtx, "data deleted successfully", "")
}

func (a *ApiCtx) storeBackup(fiberCtx *fiber.Ctx) error {
	backup, err := a.Node.FSM.BackupDB()
	if err != nil {
		return jsonresponse.ServerError(fiberCtx, "couldn't backup DB: "+err.Error())
	}

	headers := make(map[string]string)
	// Sets the filename and disposition to tell the browser there's an attachment to be downloaded.
	headers["Content-Disposition"] = "attachment; filename=backup.db"
	// Sets content type to binary file
	headers["Content-Type"] = "application/octet-stream"
	// Sets the map headers
	for k, v := range headers {
		fiberCtx.Response().Header.Set(k, v)
	}
	return fiberCtx.SendStream(bytes.NewReader(backup), len(backup))
}

func (a *ApiCtx) restoreBackup(fiberCtx *fiber.Ctx) error {
	const (
		key           = "backup"
		operationType = "RESTOREDB"
	)
	// This error handles the case when the file isn't received
	formFile, errFormFile := fiberCtx.FormFile(key)
	if errFormFile != nil {
		errMsg := fmt.Sprintf("couldn't get backup file: %v. Note: Key is '%s'", errFormFile, key)
		return jsonresponse.ServerError(fiberCtx, errMsg)
	}

	multiPartFile, errFileOpen := formFile.Open()
	if errFileOpen != nil {
		return jsonresponse.ServerError(fiberCtx, "couldn't open the received backup file: "+errFileOpen.Error())
	}
	defer multiPartFile.Close()

	buf := make([]byte, formFile.Size)
	_, errRead := multiPartFile.Read(buf)
	if errRead != nil && errRead != io.EOF {
		return jsonresponse.ServerError(fiberCtx, "couldn't read the received backup file: "+errRead.Error())
	}

	// json.RawMessage prevents non-standard types to be converted to string, and, to ensure that the unmarshalling
	// is delayed and not done in the transport.
	// This is done this way to prevent problems where non-standard json structures are double-marshalled to string
	// when they are converted/marshalled from []byte.
	payload := &fsm.Payload{
		Operation: operationType,
		Value:     json.RawMessage(buf),
	}
	errCluster := cluster.Execute(a.Node.Consensus, payload)
	if errCluster != nil {
		return jsonresponse.ServerError(fiberCtx, errCluster.Error())
	}

	return jsonresponse.OK(fiberCtx, "data restored successfully", "")
}
