package routes

import (
	"geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func AuthRoutes(app *fiber.App) {
	authGroup := app.Group("/api/v1/auth")

	authGroup.Post("/register", handlers.Register)
	authGroup.Post("/login", handlers.Login)
	authGroup.Get("/google", handlers.GoogleLogin)
	authGroup.Get("/google/callback", handlers.GoogleCallback)
}
