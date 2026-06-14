package handlers

import (
	"time"

	"geofencing_backend/database"

	"github.com/gofiber/fiber/v2"
)

func GetAlerts(c *fiber.Ctx) error {

	start := time.Now()

	geofenceID := c.Query("geofence_id")
	vehicleID := c.Query("vehicle_id")

	type AlertResponse struct {
		AlertID       string     `json:"alert_id"`
		GeofenceID    string     `json:"geofence_id"`
		GeofenceName  string     `json:"geofence_name"`
		VehicleID     *string    `json:"vehicle_id"`
		VehicleNumber *string    `json:"vehicle_number"`
		EventType     string     `json:"event_type"`
		Status        string     `json:"status"`
		CreatedAt     time.Time  `json:"created_at"`
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
		Joins("LEFT JOIN vehicles v ON v.id = a.vehicle_id")

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