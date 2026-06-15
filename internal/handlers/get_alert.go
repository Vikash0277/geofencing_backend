package handlers

import (
	"strconv"
	"time"

	"geofencing_backend/database"

	"github.com/gofiber/fiber/v2"
)

func GetAlerts(c *fiber.Ctx) error {

	start := time.Now()
	userID, err := userIDFromRequest(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	geofenceID := c.Query("geofence_id")
	vehicleID := c.Query("vehicle_id")

	type AlertResponse struct {
		AlertID       string    `json:"alert_id"`
		GeofenceID    string    `json:"geofence_id"`
		GeofenceName  string    `json:"geofence_name"`
		VehicleID     *string   `json:"vehicle_id"`
		VehicleNumber *string   `json:"vehicle_number"`
		EventType     string    `json:"event_type"`
		Status        string    `json:"status"`
		CreatedAt     time.Time `json:"created_at"`
	}

	var alerts []AlertResponse

	query := database.DB.
		Table("alert_configs a").
		Select(`
			a.id as alert_id,
			a.geofence_id,
			g.name as geofence_name,
			a.vehicle_id,
			v.vehicle_number,
			a.event_type,
			a.status,
			a.created_at
		`).
		Joins("LEFT JOIN geofences g ON g.id = a.geofence_id").
		Joins("LEFT JOIN vehicles v ON v.id = a.vehicle_id").
		Where("a.created_by = ?", userID)

	if geofenceID != "" {
		query = query.Where("a.geofence_id = ?", geofenceID)
	}

	if vehicleID != "" {
		query = query.Where("a.vehicle_id = ?", vehicleID)
	}

	if err := query.Scan(&alerts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"alerts":  alerts,
		"time_ns": time.Since(start).Nanoseconds(),
	})
}

func GetAlertEvents(c *fiber.Ctx) error {
	start := time.Now()
	userID, err := userIDFromRequest(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	limit := 50
	if rawLimit := c.Query("limit"); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed < 1 || parsed > 500 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "limit must be between 1 and 500"})
		}
		limit = parsed
	}

	type AlertEventResponse struct {
		AlertID       string    `json:"alert_id"`
		AlertConfigID string    `json:"alert_config_id"`
		GeofenceID    string    `json:"geofence_id"`
		GeofenceName  string    `json:"geofence_name"`
		VehicleID     string    `json:"vehicle_id"`
		VehicleNumber string    `json:"vehicle_number"`
		EventType     string    `json:"event_type"`
		Message       string    `json:"message"`
		Latitude      float64   `json:"latitude"`
		Longitude     float64   `json:"longitude"`
		Timestamp     time.Time `json:"timestamp"`
	}

	var events []AlertEventResponse
	query := database.DB.
		Table("alert_events ae").
		Select(`
			ae.id AS alert_id,
			ae.alert_config_id,
			ae.geofence_id,
			g.name AS geofence_name,
			ae.vehicle_id,
			v.vehicle_number,
			ae.event_type,
			ae.message,
			ae.latitude,
			ae.longitude,
			ae.timestamp
		`).
		Joins("LEFT JOIN geofences g ON g.id = ae.geofence_id").
		Joins("LEFT JOIN vehicles v ON v.id = ae.vehicle_id").
		Where("ae.user_id = ?", userID).
		Order("ae.timestamp DESC").
		Limit(limit)

	if err := query.Scan(&events).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"events":  events,
		"time_ns": time.Since(start).Nanoseconds(),
	})
}
