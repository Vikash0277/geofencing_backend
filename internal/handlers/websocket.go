package handlers

import (
	"encoding/json"
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

func AddClient(c *websocket.Conn, userID string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	clients[c] = userID
}

func RemoveClient(c *websocket.Conn) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	delete(clients, c)
}

func BroadcastAlertToUser(userID string, payload interface{}) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client, uid := range clients {
		if uid != userID {
			continue
		}
		if err := client.WriteJSON(payload); err != nil {
			log.Println("WebSocket write error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

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

func WsTrackLocationHandler(c *websocket.Conn) {
	vehicleID := c.Query("vehicle_id")
	if vehicleID == "" {
		_ = c.WriteJSON(fiber.Map{"type": "error", "error": "missing vehicle_id"})
		c.Close()
		return
	}
	defer c.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var req dto.UpdateLocationRequest
		if err := json.Unmarshal(msg, &req); err != nil {
			_ = c.WriteJSON(fiber.Map{"type": "error", "error": "invalid GPS payload"})
			continue
		}
		// The connection's vehicle ID is authoritative.
		req.VehicleID = vehicleID

		timestamp := time.Now().UTC()
		if req.Timestamp != "" {
			parsed, err := time.Parse(time.RFC3339, req.Timestamp)
			if err != nil {
				_ = c.WriteJSON(fiber.Map{"type": "error", "error": "invalid timestamp"})
				continue
			}
			timestamp = parsed
		}

		result, err := processVehicleLocation(req, timestamp)
		if err != nil {
			_ = c.WriteJSON(fiber.Map{"type": "error", "error": err.Error()})
			continue
		}

		broadcastProcessedAlerts(result)
		BroadcastAlertToAll(fiber.Map{
			"type":              "location_update",
			"vehicle_id":        req.VehicleID,
			"latitude":          req.Latitude,
			"longitude":         req.Longitude,
			"timestamp":         timestamp.Format(time.RFC3339),
			"entered_geofences": result.Entered,
			"exited_geofences":  result.Exited,
		})
		_ = c.WriteJSON(fiber.Map{
			"type":              "location_ack",
			"timestamp":         timestamp.Format(time.RFC3339),
			"entered_geofences": result.Entered,
			"exited_geofences":  result.Exited,
		})
	}
}

func WsAlertsHandler(c *websocket.Conn) {
	userID, err := userIDFromToken(c.Query("token"))
	if err != nil {
		_ = c.WriteJSON(fiber.Map{"type": "error", "error": "unauthorized"})
		c.Close()
		return
	}

	AddClient(c, userID)
	defer func() {
		RemoveClient(c)
		c.Close()
	}()

	_ = c.WriteJSON(fiber.Map{"type": "connection_ready"})
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			break
		}
	}
}
