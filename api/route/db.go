package route

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/raft"
	"github.com/narvikd/fiberparser"
	"nubedb/api/jsonresponse"
	"nubedb/consensus/fsm"
	"time"
)

func (a *ApiCtx) storeGet(fiberCtx *fiber.Ctx) error {
	payload := new(fsm.Payload)
	errParse := fiberparser.ParseAndValidate(fiberCtx, payload)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}

	value, exists, errGet := a.FSM.Get(payload.Key)
	if errGet != nil {
		return jsonresponse.ServerError(fiberCtx, "couldn't get key from DB: "+errGet.Error())
	}

	if !exists {
		return jsonresponse.NotFound(fiberCtx, "key doesn't exist")
	}
	return jsonresponse.OK(fiberCtx, "data retrieved successfully", value)
}

func (a *ApiCtx) storeSet(fiberCtx *fiber.Ctx) error {
	payload := new(fsm.Payload)
	errParse := fiberparser.ParseAndValidate(fiberCtx, payload)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}

	if a.Consensus.State() != raft.Leader {
		return jsonresponse.Make(
			fiberCtx, fiber.StatusUnprocessableEntity, false, "node is not a leader", "",
		)
	}

	payload.Operation = "SET"

	data, errMarshal := json.Marshal(&payload)
	if errMarshal != nil {
		return jsonresponse.ServerError(fiberCtx, "couldn't save data to DB: "+errMarshal.Error())
	}

	future := a.Consensus.Apply(data, 500*time.Millisecond)
	if future.Error() != nil {
		return jsonresponse.ServerError(fiberCtx, "couldn't persist data to DB Cluster: "+future.Error().Error())
	}

	response := future.Response().(*fsm.ApplyRes)
	if response.Error != nil {
		return jsonresponse.ServerError(fiberCtx, "couldn't persist data to DB Cluster. Cluster err: "+response.Error.Error())
	}

	return jsonresponse.OK(fiberCtx, "data persisted successfully", "")
}
