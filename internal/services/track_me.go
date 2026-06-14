package services

import (
	"geofencing_backend/database"
	"geofencing_backend/internal/dto"
	"geofencing_backend/internal/models"
)

func CreateTrackMe(req dto.CreateTrackMeDTO) error {

	vehicle := models.TrackMe{
		VehicleNumber: req.VehicleNumber,
		DriverName:    req.DriverName,
		VehicleType:   req.VehicleType,
		Phone:         req.Phone,
		Status:        req.Status,
	}

	if vehicle.Status == "" {
		vehicle.Status = "active"
	}

	return database.DB.Create(&vehicle).Error
}

func DeleteTrackMe(id string) error {
	return database.DB.Delete(&models.TrackMe{}, "id = ?", id).Error
}