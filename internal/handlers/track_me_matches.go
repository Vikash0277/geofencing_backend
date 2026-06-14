package handlers

import (
	"time"

	"geofencing_backend/database"

	"github.com/gofiber/fiber/v2"
)

// GetTrackMeMatches returns all trackme entries that have a matching vehicle
// (same vehicle_number OR same phone). Each matched row includes the vehicle's UUID
// so the Flutter app can connect to /ws/track?vehicle_id=<id>.
func GetTrackMeMatches(c *fiber.Ctx) error {

	start := time.Now()

	type MatchedRow struct {
		// TrackMe fields
		TrackMeID     string `gorm:"column:trackme_id" json:"trackme_id"`
		VehicleNumber string `json:"vehicle_number"`
		DriverName    string `json:"driver_name"`
		VehicleType   string `json:"vehicle_type"`
		Phone         string `json:"phone"`
		TrackMeStatus string `gorm:"column:trackme_status" json:"trackme_status"`

		// Matched Vehicle fields
		VehicleID string `gorm:"column:vehicle_id" json:"vehicle_id"` // UUID for WS tracking
	}

	var rows []MatchedRow

	err := database.DB.Raw(`
		SELECT
			t.id          AS trackme_id,
			t.vehicle_number,
			t.driver_name,
			t.vehicle_type,
			t.phone,
			t.status      AS trackme_status,
			v.id          AS vehicle_id
		FROM track_mes t
		INNER JOIN vehicles v
			ON v.vehicle_number = t.vehicle_number
			OR v.phone = t.phone
		ORDER BY t.created_at DESC
	`).Scan(&rows).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"matches": rows,
		"count":   len(rows),
		"time_ns": time.Since(start).Nanoseconds(),
	})
}
