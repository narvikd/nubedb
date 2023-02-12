package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hashicorp/raft"
	"github.com/narvikd/fiberparser"
	"log"
	"nubedb/consensus"
	"nubedb/internal/config"
	"sync"
	"time"
)

// App is a simple struct to include a collection of tools that the application could need to operate.
//
// This way the application can avoid the use of global variables.
type App struct {
	HttpServer *fiber.App
	Consensus  *raft.Raft
	DB         *sync.Map
}

func NewApp(cfg config.Config) *App {
	db := &sync.Map{}

	consen, errConsensus := consensus.New(db, cfg.ID, cfg.ConsensusAddress)
	if errConsensus != nil {
		log.Fatalln(errConsensus)
	}

	serv := fiber.New(fiber.Config{
		AppName:           "Consensus Experiment",
		EnablePrintRoutes: false,
		IdleTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ErrorHandler: func(ctx *fiber.Ctx, err error) error {
			return fiberparser.RegisterErrorHandler(ctx, err)
		},
		BodyLimit: 5 * 1024 * 1024, // In MB
	})

	return &App{
		HttpServer: serv,
		Consensus:  consen,
		DB:         db,
	}
}
