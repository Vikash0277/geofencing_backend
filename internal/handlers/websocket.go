package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	"geofencing_backend/internal/dto"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
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
		if req.VehicleID == "" {
			req.VehicleID = vehicleID
		}

		result, err := processVehicleLocation(req, time.Now())
		if err != nil {
			if !errors.Is(err, ErrStaleLocation) {
				log.Println("Failed to process WS location:", err)
			}
			continue
		}

		go broadcastProcessedAlerts(result)

		BroadcastAlertToAll(fiber.Map{
			"type":       "location_update",
			"vehicle_id": req.VehicleID,
			"latitude":   req.Latitude,
			"longitude":  req.Longitude,
			"timestamp":  time.Now().Format(time.RFC3339),
		})
	}
}
// WsAlertsHandler handles incoming websocket connections on /ws/alerts
func WsAlertsHandler(c *websocket.Conn) {
	token := c.Query("token")
	if token == "" {
		log.Println("WebSocket connection rejected: missing token")
		c.Close()
		return
	}

	userID, err := userIDFromToken(token)
	if err != nil {
		log.Println("WebSocket connection rejected: invalid token")
		c.Close()
		return
	}

	AddClient(c, userID)
	defer func() {
		RemoveClient(c)
		c.Close()
	}()

	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
