package handlers

import (
	"time"

	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"
	"geofencing_backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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

	// 1. Find previous location to determine entry/exit
	var lastLocation models.VehicleLocation
	database.DB.Where("vehicle_id = ?", req.VehicleID).Order("timestamp desc").First(&lastLocation)

	var lastGeofences []string
	if lastLocation.ID != uuid.Nil {
		database.DB.Raw(`
			SELECT id FROM geofences
			WHERE ST_Contains(
				ST_GeomFromText(polygon,4326),
				ST_SetSRID(ST_Point(?,?),4326)
			)
		`, lastLocation.Longitude, lastLocation.Latitude).Scan(&lastGeofences)
	}

	point := services.BuildPointWKT(
		req.Latitude,
		req.Longitude,
	)

	location := models.VehicleLocation{
		VehicleID: req.VehicleID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Timestamp: timestamp,
		Point:     point,
	}

	if err := database.DB.Create(&location).Error; err != nil {
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
	`, req.Longitude, req.Latitude).Scan(&geofences)

	var current []fiber.Map
	currentGeofencesMap := make(map[string]bool)

	for _, g := range geofences {
		current = append(current, fiber.Map{
			"geofence_id":   g.ID,
			"geofence_name": g.Name,
			"status":        "inside",
		})
		currentGeofencesMap[g.ID] = true
	}

	// 2. Evaluate entry and exit
	lastGeofencesMap := make(map[string]bool)
	for _, id := range lastGeofences {
		lastGeofencesMap[id] = true
	}

	var enteredGeofences []string
	var exitedGeofences []string

	for id := range currentGeofencesMap {
		if !lastGeofencesMap[id] {
			enteredGeofences = append(enteredGeofences, id)
		}
	}
	for id := range lastGeofencesMap {
		if !currentGeofencesMap[id] {
			exitedGeofences = append(exitedGeofences, id)
		}
	}

	// 3. Find and trigger alerts
	allAffected := append(enteredGeofences, exitedGeofences...)
	if len(allAffected) > 0 {
		var configs []models.AlertConfig
		database.DB.Where("geofence_id IN ? AND status = 'active'", allAffected).Find(&configs)

		for _, cfg := range configs {
			if cfg.VehicleID != nil && *cfg.VehicleID != req.VehicleID {
				continue
			}

			isEntry := false
			isExit := false
			for _, id := range enteredGeofences {
				if cfg.GeofenceID == id { isEntry = true; break }
			}
			for _, id := range exitedGeofences {
				if cfg.GeofenceID == id { isExit = true; break }
			}

			triggeredEventType := ""
			if isEntry && (cfg.EventType == "entry" || cfg.EventType == "both") {
				triggeredEventType = "entry"
			} else if isExit && (cfg.EventType == "exit" || cfg.EventType == "both") {
				triggeredEventType = "exit"
			}

			if triggeredEventType != "" {
				alertResponse := fiber.Map{
					"alert_id":    cfg.ID,
					"geofence_id": cfg.GeofenceID,
					"vehicle_id":  req.VehicleID,
					"event_type":  triggeredEventType,
					"timestamp":   time.Now().Format(time.RFC3339),
					"message":     "Vehicle " + req.VehicleID + " triggered " + triggeredEventType + " for geofence " + cfg.GeofenceID,
				}
				BroadcastAlertToUser(cfg.CreatedBy.String(), alertResponse)
			}
		}
	}

	response := fiber.Map{
		"vehicle_id":        req.VehicleID,
		"location_updated":  true,
		"current_geofences": current,
		"time_ns":           time.Since(start).Nanoseconds(),
	}

	return c.JSON(response)
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