package models

import (
	"time"

	"github.com/google/uuid"
)

type AlertEvent struct {
	BaseModel

	AlertConfigID uuid.UUID `gorm:"type:uuid;not null;index"`
	GeofenceID    string    `gorm:"type:uuid;not null;index"`
	VehicleID     string    `gorm:"type:uuid;not null;index"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index"`

	EventType string `gorm:"type:varchar(20);not null"`
	Message   string `gorm:"type:varchar(255);not null"`

	Latitude  float64   `gorm:"not null"`
	Longitude float64   `gorm:"not null"`
	Timestamp time.Time `gorm:"not null;index"`
}
