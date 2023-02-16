package route

import (
	"github.com/gofiber/fiber/v2"
)

func (a *ApiCtx) healthCheck(fiberCtx *fiber.Ctx) error {
	if !a.Node.IsHealthy() {
		return fiberCtx.Status(fiber.StatusInternalServerError).SendString("")
	}
	return fiberCtx.Status(fiber.StatusOK).SendString("")
}
