package handlers

import (
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/services"

	"github.com/gofiber/fiber/v2"
)

func CreateTrackMe(c *fiber.Ctx) error {

	var req dto.CreateTrackMeDTO

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if err := services.CreateTrackMe(req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Vehicle created successfully",
	})
}

func DeleteTrackMe(c *fiber.Ctx) error {

	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing ID parameter",
		})
	}

	if err := services.DeleteTrackMe(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Vehicle deleted successfully",
	})
}