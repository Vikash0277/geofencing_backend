package handlers

import (
	"errors"
	"time"

	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func UpdateVehicleLocation(c *fiber.Ctx) error {

	start := time.Now()

	var req dto.UpdateLocationRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid timestamp",
		})
	}

	result, err := processVehicleLocation(req, timestamp)
	if err != nil {
		if errors.Is(err, ErrStaleLocation) {
			return c.Status(409).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	go broadcastProcessedAlerts(result)

	current := make([]fiber.Map, 0, len(result.Current))
	for _, g := range result.Current {
		current = append(current, fiber.Map{
			"geofence_id":   g.ID,
			"geofence_name": g.Name,
			"status":        "inside",
		})
	}

	return c.JSON(fiber.Map{
		"vehicle_id":        req.VehicleID,
		"location_updated":  true,
		"current_geofences": current,
		"time_ns":           time.Since(start).Nanoseconds(),
	})
}



func GetVehicleLocation(c *fiber.Ctx) error {

	start := time.Now()

	vehicleID := c.Params("vehicle_id")

	var vehicle models.Vehicle

	if err := database.DB.
		First(&vehicle, "id = ?", vehicleID).
		Error; err != nil {

		return c.Status(404).JSON(fiber.Map{
			"error": "vehicle not found",
		})
	}

	var location models.VehicleLocation

	if err := database.DB.
		Order("timestamp desc").
		First(&location, "vehicle_id = ?", vehicleID).
		Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"error": "location not found",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	type GeofenceResult struct {
		ID       string
		Name     string
		Category string
	}

	var geofences []GeofenceResult

	database.DB.Raw(`
		SELECT
			id,
			name,
			category
		FROM geofences
		WHERE ST_Contains(
			ST_GeomFromText(polygon,4326),
			ST_SetSRID(ST_Point(?,?),4326)
		)
	`,
		location.Longitude,
		location.Latitude,
	).Scan(&geofences)

	var current []fiber.Map

	for _, g := range geofences {
		current = append(current, fiber.Map{
			"geofence_id":   g.ID,
			"geofence_name": g.Name,
			"category":      g.Category,
		})
	}

	return c.JSON(fiber.Map{
		"vehicle_id":     vehicle.ID,
		"vehicle_number": vehicle.VehicleNumber,

		"current_location": fiber.Map{
			"latitude":  location.Latitude,
			"longitude": location.Longitude,
			"timestamp": location.Timestamp,
		},

		"current_geofences": current,

		"time_ns": time.Since(start).Nanoseconds(),
	})
}