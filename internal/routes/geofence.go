package routes


import (
	geofenceHandler "geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func GeofenceRoutes(app *fiber.App) {
	api := app.Group("/api/v1/geofences")

	api.Post("/", geofenceHandler.CreateGeofence)
	api.Get("/", geofenceHandler.GetGeofences)
	api.Delete("/:id", geofenceHandler.DeleteGeofence)
}