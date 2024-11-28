package router

import (
	"github.com/biosecret/go-iot/handlers"
	"github.com/biosecret/go-iot/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/health", handlers.HandleHealthCheck)

	auth := app.Group("/auth")
	auth.Post("/register", handlers.RegisterHandler)
	auth.Post("/login", handlers.LoginHandler)

	api := app.Group("/api", middleware.JWTMiddleware)

	api.Get("/todos", handlers.HandleAllTodos)
	api.Post("/todos", handlers.HandleCreateTodo)
	api.Get("/todos/:id", handlers.HandleGetOneTodo)
	api.Put("/todos/:id", handlers.HandleUpdateTodo)
	api.Delete("/todos/:id", handlers.HandleDeleteTodo)
	app.Get("/sse", handlers.HandleSSE)

}
