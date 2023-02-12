package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/raft"
	"nubedb/internal/app"
	"sync"
)

// ApiCtx is a simple struct to include a collection of tools that a route could need to operate, for example a DB.
type ApiCtx struct {
	HttpServer *fiber.App
	Consensus  *raft.Raft
	DB         *sync.Map
}

// newRouteCtx returns a pointer of a new instance of ApiCtx.
func newRouteCtx(app *app.App) *ApiCtx {
	routeCtx := ApiCtx{
		HttpServer: app.HttpServer,
		Consensus:  app.Consensus,
		DB:         app.DB,
	}
	return &routeCtx
}

// Register registers fiber's routes.
func Register(app *app.App) {
	routes(app.HttpServer, newRouteCtx(app))
}

func routes(app *fiber.App, route *ApiCtx) {
	app.Get("/store", route.storeGet)
	app.Post("/store", route.storeSet)

	app.Get("/consensus", route.consensusState)
	app.Post("/consensus", route.consensusJoin)
}
