package route

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/raft"
	"nubedb/cluster/consensus/fsm"
	"nubedb/internal/app"
)

// ApiCtx is a simple struct to include a collection of tools that a route could need to operate, for example a DB.
type ApiCtx struct {
	HttpServer *fiber.App
	Consensus  *raft.Raft
	FSM        *fsm.DatabaseFSM
}

// newRouteCtx returns a pointer of a new instance of ApiCtx.
func newRouteCtx(app *app.App) *ApiCtx {
	routeCtx := ApiCtx{
		HttpServer: app.HttpServer,
		Consensus:  app.Node.Consensus,
		FSM:        app.Node.FSM,
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
	app.Delete("/store", route.storeDelete)

	app.Get("/consensus", route.consensusState)
	app.Post("/consensus", route.consensusJoin)
}
