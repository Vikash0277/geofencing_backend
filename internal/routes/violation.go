package routes

import (
	"geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func ViolationRoutes(app *fiber.App) {

	api := app.Group("/api/v1")

	api.Get("/violations/history",handlers.GetViolationHistory)
}