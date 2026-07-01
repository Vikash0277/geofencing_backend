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

	userID, err := userIDFromRequest(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req dto.ConfigureAlertRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	switch req.EventType {
	case "entry", "exit", "both":
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "event_type must be entry, exit, or both",
		})
	}

	var geofence models.Geofence
	if err := database.DB.
		First(&geofence, "id = ?", req.GeofenceID).
		Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "geofence not found",
		})
	}

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

	createdBy, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid user ID"})
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

	start := time.Now()

	userID, err := userIDFromRequest(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing ID parameter",
		})
	}

	if err := database.DB.
		Where("id = ? AND created_by = ?", id, userID).
		Delete(&models.AlertConfig{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Alert deleted successfully",
		"time_ns": time.Since(start).Nanoseconds(),
	})
}



