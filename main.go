package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"nubedb/api/proto/protoserver"
	"nubedb/api/rest/middleware"
	"nubedb/api/rest/route"
	"nubedb/internal/app"
	"nubedb/internal/config"
	"runtime"
	"sync"
)

func init() {
	if runtime.GOOS == "windows" {
		log.Fatalln("nubedb is only compatible with Mac and Linux")
	}
}

func main() {
	cfg, errCfg := config.New()
	if errCfg != nil {
		log.Fatalln(errCfg)
	}

	start(app.NewApp(cfg))
}

func start(a *app.App) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		startApiRest(a)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		startApiProto(a)
	}()

	wg.Wait()
}

func startApiProto(a *app.App) {
	log.Println("[proto] Starting proto server...")
	err := protoserver.Start(a)
	if err != nil {
		log.Fatalln("proto api can't be started:", err)
	}
}

func startApiRest(a *app.App) {
	errListen := newApiRest(a).Listen(a.Config.CurrentNode.ApiAddress)
	if errListen != nil {
		log.Fatalln("api can't be started:", errListen)
	}
}

func newApiRest(a *app.App) *fiber.App {
	middleware.InitMiddlewares(a.HttpServer)
	route.Register(a)
	return a.HttpServer
}
