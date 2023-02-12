package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"nubedb/api/middleware"
	"nubedb/api/route"
	"nubedb/internal/app"
	"nubedb/internal/config"
)

func main() {
	cfg := config.New()
	apiAddr := fmt.Sprintf("%s:%v", "0.0.0.0", cfg.ApiPort)

	errListen := newApi(cfg).Listen(apiAddr)
	if errListen != nil {
		log.Fatalln("api can't be started:", errListen)
	}
}

func newApi(cfg config.Config) *fiber.App {
	a := app.NewApp(cfg)
	middleware.InitMiddlewares(a.HttpServer)
	route.Register(a)
	return a.HttpServer
}
