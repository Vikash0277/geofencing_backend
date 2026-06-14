package main

import (
	"log"

	"geofencing_backend/database"
	"geofencing_backend/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	// Connect Database
	database.ConnectDatabase()

	// Create Fiber App
	app := fiber.New(fiber.Config{
		AppName: "Geofencing Backend",
	})

	// Enable CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, https://geofencing-frontend-phi.vercel.app",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Health Check
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Welcome to Geofencing Project",
			"status":  "running",
		})
	})

	// Routes
	routes.SetupRoutes(app)
	routes.AuthRoutes(app)
	routes.GeofenceRoutes(app)
	routes.VehicleRoutes(app)
	routes.ConfigureAlertRoutes(app)
	routes.ViolationRoutes(app)
	routes.GetAlertsRoutes(app)
	routes.TrackMeRoutes(app)

	port := ":3001"

	log.Printf("🚀 Server starting on http://localhost%s\n", port)

	// Start Server
	if err := app.Listen(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}