package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/narvikd/fiberparser"
	"log"
	"nubedb/cluster/consensus"
	"nubedb/discover"
	"nubedb/internal/config"
	"time"
)

// App is a simple struct to include a collection of tools that the application could need to operate.
//
// This way the application can avoid the use of global variables.
type App struct {
	HttpServer *fiber.App
	Node       *consensus.Node
	Config     config.Config
}

func NewApp(cfg config.Config) *App {
	errDiscover := discover.SetMode(cfg.Cluster.DiscoverStrategy)
	if errDiscover != nil {
		log.Fatalln(errDiscover)
	}

	node, errConsensus := consensus.New(cfg)
	if errConsensus != nil {
		log.Fatalln(errConsensus)
	}

	serv := fiber.New(fiber.Config{
		AppName:           "NubeDB",
		EnablePrintRoutes: false,
		IdleTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return fiberparser.RegisterErrorHandler(ctx, err)
		},
		BodyLimit: 200 * 1024 * 1024, // In MB
	})

	return &App{
		HttpServer: serv,
		Node:       node,
		Config:     cfg,
	}
}
