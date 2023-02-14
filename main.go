package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"nubedb/api/proto/protoserver"
	"nubedb/api/rest/middleware"
	"nubedb/api/rest/route"
	"nubedb/internal/app"
	"nubedb/internal/config"
	"sync"
)

func main() {
	start(config.New())
}

func start(cfg config.Config) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startApiRest(cfg)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startApiProto(cfg)
	}()

	wg.Wait()
}

func startApiProto(cfg config.Config) {
	log.Println("Starting proto server...")
	err := protoserver.Start(cfg)
	if err != nil {
		log.Fatalln(err)
	}
}

func startApiRest(cfg config.Config) {
	errListen := newApiRest(cfg).Listen(cfg.ApiAddress)
	if errListen != nil {
		log.Fatalln("api can't be started:", errListen)
	}
}

func newApiRest(cfg config.Config) *fiber.App {
	a := app.NewApp(cfg)
	middleware.InitMiddlewares(a.HttpServer)
	route.Register(a)
	return a.HttpServer
}
