package routes

import (
	"geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {

	api := app.Group("/api/v1")

	api.Post("/vehicles/location", handlers.UpdateVehicleLocation)

	api.Get("/vehicles/location/:vehicle_id", handlers.GetVehicleLocation)
}