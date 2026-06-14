package routes

import (
	"geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func ConfigureAlertRoutes(app *fiber.App) {

	api := app.Group("/api/v1")

	api.Post(
		"/alerts/configure",
		handlers.ConfigureAlert,
	)

	api.Get(
		"/alerts",
		handlers.GetAlerts,
	)

	api.Delete(
		"/alerts/:id",
		handlers.DeleteAlert,
	)
}
