package routes

import (
	"geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func VehicleRoutes(app *fiber.App) {

	api := app.Group("/api/v1")

	api.Post("/vehicles", handlers.CreateVehicle)
	api.Get("/vehicles", handlers.GetVehicles)
	api.Delete("/vehicles/:id", handlers.DeleteVehicle)
}