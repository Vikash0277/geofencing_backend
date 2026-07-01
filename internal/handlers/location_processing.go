package handlers

import (
	"errors"
	"fmt"
	"time"

	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"
	"geofencing_backend/internal/services"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrStaleLocation = errors.New("location timestamp is not newer than the last location")

type currentGeofence struct {
	ID       string
	Name     string
	Category string
}

type geofenceAlertPayload struct {
	Type      string `json:"type"`
	EventID   string `json:"event_id"`
	EventType string `json:"event_type"`
	Timestamp string `json:"timestamp"`
	UserID    string `json:"-"`
	Vehicle   struct {
		VehicleID      string `json:"vehicle_id"`
		VehicleNumber  string `json:"vehicle_number"`
		DriverName     string `json:"driver_name"`
	} `json:"vehicle"`
	Geofence struct {
		GeofenceID   string `json:"geofence_id"`
		GeofenceName string `json:"geofence_name"`
		Category     string `json:"category"`
	} `json:"geofence"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
}

type locationProcessingResult struct {
	Current []currentGeofence
	Entered []string
	Exited  []string
	Alerts  []geofenceAlertPayload
}

func validateLocationRequest(req dto.UpdateLocationRequest) error {
	if _, err := uuid.Parse(req.VehicleID); err != nil {
		return errors.New("invalid vehicle_id")
	}
	if req.Latitude < -90 || req.Latitude > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}

func geofencesAt(tx *gorm.DB, latitude, longitude float64) ([]currentGeofence, error) {
	var geofences []currentGeofence
	err := tx.Raw(`
		SELECT id, name, category
		FROM geofences
		WHERE status = 'active'
		  AND ST_Covers(
			polygon,
			ST_SetSRID(ST_Point(?, ?), 4326)
		  )
	`, longitude, latitude).Scan(&geofences).Error
	return geofences, err
}

func processVehicleLocation(req dto.UpdateLocationRequest, timestamp time.Time) (locationProcessingResult, error) {
	result := locationProcessingResult{
		Current: make([]currentGeofence, 0),
		Entered: make([]string, 0),
		Exited:  make([]string, 0),
		Alerts:  make([]geofenceAlertPayload, 0),
	}

	if err := validateLocationRequest(req); err != nil {
		return result, err
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Serialize updates for one vehicle so concurrent GPS frames cannot emit duplicates.
		var lockResult int
		if err := tx.Raw(
			"SELECT pg_advisory_xact_lock(hashtext(?)), 1",
			req.VehicleID,
		).Row().Scan(new(interface{}), &lockResult); err != nil {
			return err
		}

		var vehicle models.Vehicle
		if err := tx.First(&vehicle, "id = ?", req.VehicleID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("vehicle not found")
			}
			return err
		}

		var lastLocation models.VehicleLocation
		lastLocationErr := tx.
			Where("vehicle_id = ?", req.VehicleID).
			Order("timestamp DESC").
			First(&lastLocation).Error
		if lastLocationErr != nil && !errors.Is(lastLocationErr, gorm.ErrRecordNotFound) {
			return lastLocationErr
		}
		if lastLocationErr == nil && !timestamp.After(lastLocation.Timestamp) {
			return ErrStaleLocation
		}

		previous := make([]currentGeofence, 0)
		if lastLocationErr == nil {
			var err error
			previous, err = geofencesAt(tx, lastLocation.Latitude, lastLocation.Longitude)
			if err != nil {
				return err
			}
		}

		var err error
		result.Current, err = geofencesAt(tx, req.Latitude, req.Longitude)
		if err != nil {
			return err
		}

		location := models.VehicleLocation{
			VehicleID: req.VehicleID,
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Timestamp: timestamp,
			Point:     services.BuildPointWKT(req.Latitude, req.Longitude),
		}
		if err := tx.Create(&location).Error; err != nil {
			return err
		}

		previousMap := make(map[string]bool, len(previous))
		currentMap := make(map[string]bool, len(result.Current))
		for _, geofence := range previous {
			previousMap[geofence.ID] = true
		}
		for _, geofence := range result.Current {
			currentMap[geofence.ID] = true
			if !previousMap[geofence.ID] {
				result.Entered = append(result.Entered, geofence.ID)
			}
		}
		for _, geofence := range previous {
			if !currentMap[geofence.ID] {
				result.Exited = append(result.Exited, geofence.ID)
			}
		}

		events := make(map[string]string, len(result.Entered)+len(result.Exited))
		for _, id := range result.Entered {
			events[id] = "entry"
		}
		for _, id := range result.Exited {
			events[id] = "exit"
		}
		if len(events) == 0 {
			return nil
		}

		affected := make([]string, 0, len(events))
		for id := range events {
			affected = append(affected, id)
		}

		type geofenceInfo struct {
			ID        string
			Name      string
			Category  string
			CreatedBy string
		}
		var gfInfos []geofenceInfo
		if err := tx.Table("geofences").
			Select("id, name, category, created_by").
			Where("id IN ?", affected).
			Scan(&gfInfos).Error; err != nil {
			return err
		}
		gfInfoMap := make(map[string]geofenceInfo, len(gfInfos))
		for _, info := range gfInfos {
			gfInfoMap[info.ID] = info
		}

		for geofenceID, eventType := range events {
			gfInfo := gfInfoMap[geofenceID]
			violation := models.Violation{
				VehicleID:  req.VehicleID,
				GeofenceID: geofenceID,
				UserID:     gfInfo.CreatedBy,
				EventType:  eventType,
				Latitude:   req.Latitude,
				Longitude:  req.Longitude,
				Timestamp:  timestamp,
			}
			if err := tx.Create(&violation).Error; err != nil {
				return err
			}
		}

		var configs []models.AlertConfig
		if err := tx.
			Where("geofence_id IN ? AND status = 'active'", affected).
			Find(&configs).Error; err != nil {
			return err
		}

		for _, config := range configs {
			eventType := events[config.GeofenceID]
			if config.VehicleID != nil && *config.VehicleID != req.VehicleID {
				continue
			}
			if config.EventType != eventType && config.EventType != "both" {
				continue
			}

			gfInfo := gfInfoMap[config.GeofenceID]

			message := fmt.Sprintf(
				"Vehicle %s triggered %s for geofence %s",
				vehicle.VehicleNumber,
				eventType,
				gfInfo.Name,
			)

			event := models.AlertEvent{
				AlertConfigID: config.ID.String(),
				GeofenceID:    config.GeofenceID,
				VehicleID:     req.VehicleID,
				UserID:        config.CreatedBy.String(),
				EventType:     eventType,
				Message:       message,
				Latitude:      req.Latitude,
				Longitude:     req.Longitude,
				Timestamp:     timestamp,
			}
			if err := tx.Create(&event).Error; err != nil {
				return err
			}

			result.Alerts = append(result.Alerts, geofenceAlertPayload{
				Type:      "geofence_alert",
				EventID:   event.ID.String(),
				EventType: eventType,
				Timestamp: timestamp.Format(time.RFC3339),
				UserID:    config.CreatedBy.String(),
				Vehicle: struct {
					VehicleID      string `json:"vehicle_id"`
					VehicleNumber  string `json:"vehicle_number"`
					DriverName     string `json:"driver_name"`
				}{
					VehicleID:      req.VehicleID,
					VehicleNumber:  vehicle.VehicleNumber,
					DriverName:     vehicle.DriverName,
				},
				Geofence: struct {
					GeofenceID   string `json:"geofence_id"`
					GeofenceName string `json:"geofence_name"`
					Category     string `json:"category"`
				}{
					GeofenceID:   config.GeofenceID,
					GeofenceName: gfInfo.Name,
					Category:     gfInfo.Category,
				},
				Location: struct {
					Latitude  float64 `json:"latitude"`
					Longitude float64 `json:"longitude"`
				}{
					Latitude:  req.Latitude,
					Longitude: req.Longitude,
				},
			})
		}

		return nil
	})

	return result, err
}

func broadcastProcessedAlerts(result locationProcessingResult) {
	for _, alert := range result.Alerts {
		BroadcastAlertToUser(alert.UserID, alert)
	}
}
