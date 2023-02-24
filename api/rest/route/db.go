package route

import (
	"bytes"
	"github.com/gofiber/fiber/v2"
	"github.com/narvikd/fiberparser"
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
	// set filename and disposition
	headers["Content-Disposition"] = "attachment; filename=backup.db"
	// set content type to binary
	headers["Content-Type"] = "application/octet-stream"
	// set the headers
	for k, v := range headers {
		fiberCtx.Response().Header.Set(k, v)
	}
	return fiberCtx.SendStream(bytes.NewReader(backup), len(backup))
}
