package jsonresponse

import "github.com/gofiber/fiber/v2"

func Make(ctx *fiber.Ctx, status int, success bool, message string, data interface{}) error {
	return ctx.Status(status).JSON(&fiber.Map{
		"success": success,
		"message": message,
		"data":    data,
	})
}

// OK returns a successful response with status code 200
func OK(ctx *fiber.Ctx, message string, data interface{}) error {
	return ctx.Status(200).JSON(&fiber.Map{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// NotFound returns a not found response with status code 404
func NotFound(ctx *fiber.Ctx, message string) error {
	return ctx.Status(404).JSON(&fiber.Map{
		"success": false,
		"message": message,
		"data":    "",
	})
}

// BadRequest returns a bad request response with status code http status 400
func BadRequest(ctx *fiber.Ctx, message string) error {
	return ctx.Status(400).JSON(&fiber.Map{
		"success": false,
		"message": message,
		"data":    "",
	})
}

// ServerError returns a server error response with status code 500
func ServerError(ctx *fiber.Ctx, message string) error {
	return ctx.Status(500).JSON(&fiber.Map{
		"success": false,
		"message": message,
		"data":    "",
	})
}
