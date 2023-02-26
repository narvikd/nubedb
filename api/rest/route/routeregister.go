package route

import (
	"github.com/gofiber/fiber/v2"
	"nubedb/cluster/consensus"
	"nubedb/internal/app"
	"nubedb/internal/config"
)

// ApiCtx is a simple struct to include a collection of tools that a route could need to operate, for example a DB.
type ApiCtx struct {
	Config     config.Config
	HttpServer *fiber.App
	Node       *consensus.Node
}

// newRouteCtx returns a pointer of a new instance of ApiCtx.
func newRouteCtx(app *app.App) *ApiCtx {
	routeCtx := ApiCtx{
		Config:     app.Config,
		HttpServer: app.HttpServer,
		Node:       app.Node,
	}
	return &routeCtx
}

// Register registers fiber's routes.
func Register(app *app.App) {
	routes(app.HttpServer, newRouteCtx(app))
}

func routes(app *fiber.App, route *ApiCtx) {
	app.Get("/store", route.storeGet)
	app.Get("/store/keys", route.storeGetKeys)

	app.Post("/store", route.storeSet)
	app.Delete("/store", route.storeDelete)

	app.Get("/store/backup", route.storeBackup)
	app.Post("/store/restore", route.restoreBackup)

	app.Get("/consensus", route.consensusState)
	app.Get("/healthcheck", route.healthCheck)
}
