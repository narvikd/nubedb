package route

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/raft"
	"github.com/narvikd/errorskit"
	"github.com/narvikd/fiberparser"
	"nubedb/api/rest/jsonresponse"
	"nubedb/cluster/consensus"
)

func (a *ApiCtx) consensusState(fiberCtx *fiber.Ctx) error {
	stats := a.Node.Consensus.Stats()
	address, id := a.Node.Consensus.LeaderWithID()
	stats["leader"] = fmt.Sprintf("Address: %s Leader ID: %s", address, id)
	return jsonresponse.OK(fiberCtx, "consensus state retrieved successfully", stats)
}

func (a *ApiCtx) consensusJoin(fiberCtx *fiber.Ctx) error {
	c := new(consensus.Node)
	errParse := fiberparser.ParseAndValidate(fiberCtx, c)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}

	if c.Address == "" {
		return jsonresponse.BadRequest(fiberCtx, "address can't be empty")
	}

	if a.Node.Consensus.State() != raft.Leader {
		return jsonresponse.BadRequest(fiberCtx, "node is not a leader")
	}

	future := a.Node.Consensus.AddVoter(raft.ServerID(c.ID), raft.ServerAddress(c.Address), 0, 0)
	if future.Error() != nil {
		return jsonresponse.ServerError(fiberCtx, errorskit.Wrap(future.Error(), "failed to add server").Error())
	}

	return jsonresponse.OK(fiberCtx, fmt.Sprintf("node '%s' added to consensus", c.ID), "")
}

func (a *ApiCtx) consensusRemove(fiberCtx *fiber.Ctx) error {
	c := new(consensus.Node)
	errParse := fiberparser.ParseAndValidate(fiberCtx, c)
	if errParse != nil {
		return jsonresponse.BadRequest(fiberCtx, errParse.Error())
	}

	if a.Node.Consensus.State() != raft.Leader {
		return jsonresponse.BadRequest(fiberCtx, "node is not a leader")
	}

	future := a.Node.Consensus.RemoveServer(raft.ServerID(c.ID), 0, 0)
	if future.Error() != nil {
		return jsonresponse.ServerError(fiberCtx, errorskit.Wrap(future.Error(), "failed to remove server").Error())
	}

	return jsonresponse.OK(fiberCtx, fmt.Sprintf("node '%s' removed from consensus", c.ID), "")
}
