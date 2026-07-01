package handlers

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"
	"geofencing_backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

var (
	clients   = make(map[*websocket.Conn]string)
	clientsMu sync.Mutex
)

// AddClient registers a new websocket client for a specific user
func AddClient(c *websocket.Conn, userID string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[c] = userID
}

// RemoveClient removes a websocket client
func RemoveClient(c *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, c)
}

// BroadcastAlertToUser sends a payload to all connected websocket clients of a specific user
func BroadcastAlertToUser(userID string, payload interface{}) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client, uid := range clients {
		if uid == userID {
			if err := client.WriteJSON(payload); err != nil {
				log.Println("WebSocket write error:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// BroadcastAlertToAll sends a payload to every connected websocket client (used for live location updates)
func BroadcastAlertToAll(payload interface{}) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if err := client.WriteJSON(payload); err != nil {
			log.Println("WebSocket broadcast error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// WsTrackLocationHandler is the WebSocket endpoint Flutter connects to.
// Flutter sends: {"vehicle_id": "...", "latitude": 0.0, "longitude": 0.0, "timestamp": "..."}
// The handler saves the location, runs PostGIS geofencing checks, fires alerts,
// and broadcasts a location_update to all connected frontend clients.
func WsTrackLocationHandler(c *websocket.Conn) {
	vehicleID := c.Query("vehicle_id")
	if vehicleID == "" {
		log.Println("WS /ws/track rejected: missing vehicle_id")
		c.Close()
		return
	}

	log.Printf("Flutter device connected for vehicle: %s", vehicleID)
	defer func() {
		log.Printf("Flutter device disconnected for vehicle: %s", vehicleID)
		c.Close()
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Flutter WS read error:", err)
			break
		}

		var req dto.UpdateLocationRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Println("Invalid GPS payload from Flutter:", err)
			continue
		}
		// Override vehicle_id from query param (device is authoritative)
		if req.VehicleID == "" {
			req.VehicleID = vehicleID
		}

		// 1. Find previous location to determine geofence entry/exit
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

		// 2. Save new location
		point := services.BuildPointWKT(req.Latitude, req.Longitude)
		location := models.VehicleLocation{
			VehicleID: req.VehicleID,
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Timestamp: time.Now(),
			Point:     point,
		}
		if err := database.DB.Create(&location).Error; err != nil {
			log.Println("Failed to save WS location:", err)
			continue
		}

		// 3. Fetch current geofences the point falls inside
		type GeofenceResult struct {
			ID   string
			Name string
		}
		var geofences []GeofenceResult
		database.DB.Raw(`
			SELECT id, name FROM geofences
			WHERE ST_Contains(ST_GeomFromText(polygon,4326), ST_SetSRID(ST_Point(?,?),4326))
		`, req.Longitude, req.Latitude).Scan(&geofences)

		currentGeofencesMap := make(map[string]bool)
		for _, g := range geofences {
			currentGeofencesMap[g.ID] = true
		}
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

		// 4. Trigger geofence alerts
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
					alertPayload := fiber.Map{
						"type":        "geofence_alert",
						"alert_id":    cfg.ID,
						"geofence_id": cfg.GeofenceID,
						"vehicle_id":  req.VehicleID,
						"event_type":  triggeredEventType,
						"timestamp":   time.Now().Format(time.RFC3339),
						"message":     "Vehicle " + req.VehicleID + " " + triggeredEventType + " geofence " + cfg.GeofenceID,
					}
					BroadcastAlertToUser(cfg.CreatedBy.String(), alertPayload)
				}
			}
		}

		// 5. Broadcast raw location to all frontend map clients
		BroadcastAlertToAll(fiber.Map{
			"type":       "location_update",
			"vehicle_id": req.VehicleID,
			"latitude":   req.Latitude,
			"longitude":  req.Longitude,
			"timestamp":  location.Timestamp.Format(time.RFC3339),
		})

		log.Printf("[WS Track] vehicle=%s lat=%.6f lng=%.6f entered=%v exited=%v",
			req.VehicleID, req.Latitude, req.Longitude, enteredGeofences, exitedGeofences)
	}
}

// WsAlertsHandler handles incoming websocket connections on /ws/alerts
func WsAlertsHandler(c *websocket.Conn) {
	userID := c.Query("user_id")
	if userID == "" {
		log.Println("WebSocket connection rejected: missing user_id")
		c.Close()
		return
	}

	AddClient(c, userID)
	defer func() {
		RemoveClient(c)
		c.Close()
	}()

	// Keep connection alive, listen for close or incoming messages
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
