package config

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	_ "github.com/biosecret/go-iot/docs"
)

func AddSwaggerRoutes(app *fiber.App) {
	// setup swagger
	app.Get("/swagger/*", swagger.HandlerDefault)
}
