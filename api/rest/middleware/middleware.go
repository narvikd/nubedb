package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// InitMiddlewares initializes/registers all the app middlewares.
func InitMiddlewares(app *fiber.App) {
	initCorsMW(app)
	initRecoverMW(app)
}

// initCorsMW is set to allow all.
func initCorsMW(app *fiber.App) {
	app.Use(
		cors.New(cors.Config{
			AllowCredentials: true,
		}),
	)
}

// initRecoverMW initializes the Recover MW.
//
// If this is active, on newApp, the error handle must be set to fiberparser.RegisterErrorHandler
// to prevent information disclosure to the client.
func initRecoverMW(app *fiber.App) {
	app.Use(recover.New())
}
