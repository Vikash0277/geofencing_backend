package handlers

import (
	//"log"
	"time"

	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ConfigureAlert(c *fiber.Ctx) error {

	start := time.Now()

	var req dto.ConfigureAlertRequest

	if err := c.BodyParser(&req); err != nil {
		//log.Printf("ConfigureAlert BodyParser error: %v, body: %s", err, string(c.Body()))
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Validate event type
	switch req.EventType {
	case "entry", "exit", "both":
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "event_type must be entry, exit, or both",
		})
	}

	// Check geofence exists
	var geofence models.Geofence

	if err := database.DB.
		First(&geofence, "id = ?", req.GeofenceID).
		Error; err != nil {

		return c.Status(404).JSON(fiber.Map{
			"error": "geofence not found",
		})
	}

	// Optional vehicle validation
	if req.VehicleID != nil {

		var vehicle models.Vehicle

		if err := database.DB.
			First(&vehicle, "id = ?", *req.VehicleID).
			Error; err != nil {

			return c.Status(404).JSON(fiber.Map{
				"error": "vehicle not found",
			})
		}
	}

	// Parse CreatedBy UUID
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid created_by UUID"})
	}

	alert := models.AlertConfig{
		GeofenceID: req.GeofenceID,
		VehicleID:  req.VehicleID,
		CreatedBy:  createdBy,
		EventType:  req.EventType,
		Status:     "active",
	}

	if err := database.DB.Create(&alert).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"alert_id":    alert.ID,
		"geofence_id": alert.GeofenceID,
		"vehicle_id":  alert.VehicleID,
		"event_type":  alert.EventType,
		"status":      alert.Status,
		"time_ns":     time.Since(start).Nanoseconds(),
	})
}

func DeleteAlert(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing ID parameter",
		})
	}

	alertID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid alert ID",
		})
	}

	result := database.DB.Delete(&models.AlertConfig{}, "id = ?", alertID)
	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": result.Error.Error(),
		})
	}
	if result.RowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "alert rule not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Alert rule deleted successfully",
	})
}
