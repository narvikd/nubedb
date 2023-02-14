package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
	"nubedb/api/rest/middleware"
	"nubedb/api/rest/route"
	"nubedb/internal/app"
	"nubedb/internal/config"
)

func main() {
	cfg := config.New()
	errListen := newApi(cfg).Listen(cfg.ApiAddress)
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
