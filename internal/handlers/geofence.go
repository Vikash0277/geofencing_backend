package handlers

import (
	"log"
	"time"

	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateGeofence(c *fiber.Ctx) error {

	var req dto.CreateGeofenceRequest

	log.Println("CreatedBy RAW:", req.CreatedBy)
	log.Println("coordinates", c.Body())


	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := services.ValidateCoordinates(req.Coordinates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	wkt := services.BuildWKT(req.Coordinates)

	// Parse CreatedBy UUID
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid created_by UUID"})
	}

	start := time.Now()

	// Insert using raw SQL with PostGIS ST_GeomFromText
	var newID string
	err = database.DB.Raw(`
		INSERT INTO geofences (id, created_at, updated_at, name, description, category, status, created_by, polygon)
		VALUES (gen_random_uuid(), NOW(), NOW(), ?, ?, ?, ?, ?, ST_GeomFromText(?, 4326))
		RETURNING id
	`, req.Name, req.Description, req.Category, "active", createdBy, wkt).Scan(&newID).Error

	if err != nil {
		log.Println("error", err.Error())
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "geofence created successfully",
		"id":      newID,
		"name":    req.Name,
		"status":  "active",
		"time_ns": time.Since(start).Nanoseconds(),
	})
}


type GeofenceRow struct {
	ID          string
	Name        string
	Description string
	Category    string
	CreatedAt   time.Time
	PolygonWKT  string
}

func GetGeofences(c *fiber.Ctx) error {

	start := time.Now()

	category := c.Query("category")

	var rows []GeofenceRow

	query := database.DB.Table("geofences").
		Select(`
			id,
			name,
			description,
			category,
			created_at,
			ST_AsText(polygon) as polygon_wkt
		`)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Find(&rows).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := make([]dto.GeofenceResponse, 0, len(rows))

	for _, r := range rows {

		coords := services.WKTToCoords(r.PolygonWKT)

		response = append(response, dto.GeofenceResponse{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Category:    r.Category,
			Coordinates: coords,
			CreatedAt:   r.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(fiber.Map{
		"geofences": response,
		"time_ns":   time.Since(start).Nanoseconds(),
	})
}

func DeleteGeofence(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing ID parameter",
		})
	}

	tx := database.DB.Begin()
	if err := tx.Exec("DELETE FROM alert_configs WHERE geofence_id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := tx.Exec("DELETE FROM violations WHERE geofence_id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := tx.Exec("DELETE FROM geofences WHERE id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	tx.Commit()

	return c.JSON(fiber.Map{
		"message": "Geofence deleted successfully",
	})
}