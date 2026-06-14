package routes

import (
	"geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func TrackMeRoutes(app *fiber.App) {

	api := app.Group("/api/v1/trackme")

	api.Post("/", handlers.CreateTrackMe)
	api.Delete("/:id", handlers.DeleteTrackMe)

	// Returns trackme rows that have a matching vehicle (same vehicle_number OR phone)
	// Flutter uses this to show the table and get vehicle_id for WS tracking
	api.Get("/matches", handlers.GetTrackMeMatches)
}