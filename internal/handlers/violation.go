package handlers

import (
	"strconv"
	"time"

	"geofencing_backend/database"

	"github.com/gofiber/fiber/v2"
)

func GetViolationHistory(c *fiber.Ctx) error {

	start := time.Now()

	vehicleID := c.Query("vehicle_id")
	geofenceID := c.Query("geofence_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	limit := 50

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			if parsed > 0 && parsed <= 500 {
				limit = parsed
			}
		}
	}

	type ViolationResponse struct {
		ID            string    `json:"id"`
		VehicleID     string    `json:"vehicle_id"`
		VehicleNumber string    `json:"vehicle_number"`
		GeofenceID    string    `json:"geofence_id"`
		GeofenceName  string    `json:"geofence_name"`
		EventType     string    `json:"event_type"`
		Latitude      float64   `json:"latitude"`
		Longitude     float64   `json:"longitude"`
		Timestamp     time.Time `json:"timestamp"`
	}

	var violations []ViolationResponse
	var totalCount int64

	query := database.DB.
		Table("violations v").
		Select(`
			v.id,
			v.vehicle_id,
			veh.vehicle_number,
			v.geofence_id,
			g.name as geofence_name,
			v.event_type,
			v.latitude,
			v.longitude,
			v.timestamp
		`).
		Joins("LEFT JOIN vehicles veh ON veh.id = v.vehicle_id").
		Joins("LEFT JOIN geofences g ON g.id = v.geofence_id")

	if vehicleID != "" {
		query = query.Where("v.vehicle_id = ?", vehicleID)
	}

	if geofenceID != "" {
		query = query.Where("v.geofence_id = ?", geofenceID)
	}

	if startDate != "" {
		query = query.Where("v.timestamp >= ?", startDate)
	}

	if endDate != "" {
		query = query.Where("v.timestamp <= ?", endDate)
	}

	query.Count(&totalCount)

	if err := query.
		Order("v.timestamp DESC").
		Limit(limit).
		Scan(&violations).Error; err != nil {

		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"violations":  violations,
		"total_count": totalCount,
		"time_ns":     time.Since(start).Nanoseconds(),
	})
}