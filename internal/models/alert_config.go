package models

import (
	"github.com/google/uuid"
)

type AlertConfig struct {
	BaseModel

	GeofenceID string `gorm:"type:uuid;not null;index"`
	VehicleID *string `gorm:"type:uuid;index"`

	CreatedBy uuid.UUID `gorm:"type:uuid;not null;index"`

	EventType string `gorm:"type:varchar(50)"`

	Status string `gorm:"default:active"`
}