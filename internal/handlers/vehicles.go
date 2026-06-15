package handlers

import (
	"time"

	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"

	"github.com/gofiber/fiber/v2"
)

func CreateVehicle(c *fiber.Ctx) error {

	start := time.Now()

	var req dto.CreateVehicleRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.VehicleNumber == "" || req.DriverName == "" ||
		req.VehicleType == "" || req.Phone == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "all fields are required",
		})
	}

	vehicle := models.Vehicle{
		VehicleNumber: req.VehicleNumber,
		DriverName:    req.DriverName,
		VehicleType:   req.VehicleType,
		Phone:         req.Phone,
		Status:        "active",
	}

	if err := database.DB.Create(&vehicle).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(dto.CreateVehicleResponse{
		ID:            vehicle.ID.String(),
		VehicleNumber: vehicle.VehicleNumber,
		Status:        vehicle.Status,
		TimeNS:        time.Since(start).Nanoseconds(),
	})
}

func GetVehicles(c *fiber.Ctx) error {

	start := time.Now()

	var vehicles []models.Vehicle

	if err := database.DB.Table("vehicles").
		Select("vehicles.*").
		Joins("INNER JOIN track_mes ON vehicles.vehicle_number = track_mes.vehicle_number OR vehicles.phone = track_mes.phone").
		Find(&vehicles).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := make([]dto.VehicleResponse, 0, len(vehicles))

	for _, v := range vehicles {
		response = append(response, dto.VehicleResponse{
			ID:            v.ID.String(),
			VehicleNumber: v.VehicleNumber,
			DriverName:    v.DriverName,
			VehicleType:   v.VehicleType,
			Phone:         v.Phone,
			Status:        v.Status,
			CreatedAt:     v.CreatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(fiber.Map{
		"vehicles": response,
		"time_ns":  time.Since(start).Nanoseconds(),
	})
}

func DeleteVehicle(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing ID parameter",
		})
	}

	tx := database.DB.Begin()
	if err := tx.Exec("DELETE FROM alert_events WHERE vehicle_id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := tx.Exec("DELETE FROM alert_configs WHERE vehicle_id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := tx.Exec("DELETE FROM violations WHERE vehicle_id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := tx.Exec("DELETE FROM vehicle_locations WHERE vehicle_id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if err := tx.Exec("DELETE FROM vehicles WHERE id = ?", id).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	tx.Commit()

	return c.JSON(fiber.Map{
		"message": "Vehicle deleted successfully",
	})
}
