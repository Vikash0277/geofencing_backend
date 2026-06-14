package models

import (
	"time"
)

type VehicleLocation struct {
	BaseModel

	VehicleID string    `gorm:"type:uuid;not null;index"`
	Vehicle   Vehicle   `gorm:"foreignKey:VehicleID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Latitude  float64   `gorm:"not null"`
	Longitude float64   `gorm:"not null"`

	Timestamp time.Time `gorm:"not null"`

	Point     string    `gorm:"type:geometry(POINT,4326)"`
}