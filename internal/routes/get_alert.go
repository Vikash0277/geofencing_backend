package routes

import (
	"geofencing_backend/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func GetAlertsRoutes(app *fiber.App) {

	// Middleware to check if the request is a websocket upgrade
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// Websocket route for alerts (frontend → receives geofence alerts + location_update frames)
	app.Get("/ws/alerts", websocket.New(handlers.WsAlertsHandler))

	// Websocket route for Flutter GPS tracking (Flutter app → sends GPS frames here)
	app.Get("/ws/track", websocket.New(handlers.WsTrackLocationHandler))

	api := app.Group("/api/v1")

	api.Get(
		"/alerts",
		handlers.GetAlerts,
	)
	api.Get("/alert-events", handlers.GetAlertEvents)
}
