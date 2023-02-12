package route

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/fiberparser"
	"nubedb/api/jsonresponse"
	"nubedb/consensus"
)

func (a *ApiCtx) consensusJoin(fiberCtx *fiber.Ctx) error {
	c := new(consensus.Node)
	errParse := fiberparser.ParseAndValidate(fiberCtx, c)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}

	if a.Consensus.State() != raft.Leader {
		return jsonresponse.BadRequest(fiberCtx, "node is not a leader")
	}

	future := a.Consensus.AddVoter(raft.ServerID(c.ID), raft.ServerAddress(c.Address), 0, 0)
	if future.Error() != nil {
		return jsonresponse.ServerError(fiberCtx, errorskit.Wrap(future.Error(), "failed to add voter").Error())
	}

	return jsonresponse.OK(fiberCtx, fmt.Sprintf("node '%s' added to consensus", c.ID), "")
}

func (a *ApiCtx) consensusState(fiberCtx *fiber.Ctx) error {
	return jsonresponse.OK(fiberCtx, "consensus state retrieved successfully", a.Consensus.Stats())
}
