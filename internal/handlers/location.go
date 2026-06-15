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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid timestamp"})
	}

	result, err := processVehicleLocation(req, timestamp)
	if err != nil {
		status := fiber.StatusInternalServerError
		if errors.Is(err, ErrStaleLocation) {
			status = fiber.StatusConflict
		} else if err.Error() == "vehicle not found" {
			status = fiber.StatusNotFound
		} else if err.Error() == "invalid vehicle_id" ||
			err.Error() == "latitude must be between -90 and 90" ||
			err.Error() == "longitude must be between -180 and 180" {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}

	broadcastProcessedAlerts(result)

	current := make([]fiber.Map, 0, len(result.Current))
	for _, geofence := range result.Current {
		current = append(current, fiber.Map{
			"geofence_id":   geofence.ID,
			"geofence_name": geofence.Name,
			"category":      geofence.Category,
			"status":        "inside",
		})
	}

	return c.JSON(fiber.Map{
		"vehicle_id":        req.VehicleID,
		"location_updated":  true,
		"current_geofences": current,
		"entered_geofences": result.Entered,
		"exited_geofences":  result.Exited,
		"triggered_alerts":  result.Alerts,
		"time_ns":           time.Since(start).Nanoseconds(),
	})
}

func GetVehicleLocation(c *fiber.Ctx) error {
	start := time.Now()
	vehicleID := c.Params("vehicle_id")

	var vehicle models.Vehicle
	if err := database.DB.First(&vehicle, "id = ?", vehicleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "vehicle not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var location models.VehicleLocation
	if err := database.DB.
		Order("timestamp DESC").
		First(&location, "vehicle_id = ?", vehicleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "location not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	geofences, err := geofencesAt(database.DB, location.Latitude, location.Longitude)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	current := make([]fiber.Map, 0, len(geofences))
	for _, geofence := range geofences {
		current = append(current, fiber.Map{
			"geofence_id":   geofence.ID,
			"geofence_name": geofence.Name,
			"category":      geofence.Category,
			"status":        "inside",
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
		"time_ns":           time.Since(start).Nanoseconds(),
	})
}
