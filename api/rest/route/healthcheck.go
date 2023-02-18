package route

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"nubedb/api/rest/jsonresponse"
)

func (a *ApiCtx) healthCheck(fiberCtx *fiber.Ctx) error {
	if !a.Node.IsHealthy() {
		return fiberCtx.Status(fiber.StatusInternalServerError).SendString("")
	}
	return fiberCtx.Status(fiber.StatusOK).SendString("")
}

func (a *ApiCtx) consensusState(fiberCtx *fiber.Ctx) error {
	stats := a.Node.Consensus.Stats()
	address, id := a.Node.Consensus.LeaderWithID()
	stats["leader"] = fmt.Sprintf("Address: %s Leader ID: %s", address, id)
	stats["node_id"] = a.Config.CurrentNode.ID
	return jsonresponse.OK(fiberCtx, "consensus state retrieved successfully", stats)
}
