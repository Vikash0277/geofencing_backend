package models

import "time"

type Violation struct {
	BaseModel

	VehicleID string `gorm:"type:uuid;not null;index"`
	GeofenceID string `gorm:"type:uuid;not null;index"`

	EventType string `gorm:"type:varchar(50)"` 

	Latitude  float64 `gorm:"not null"`
	Longitude float64 `gorm:"not null"`

	Timestamp time.Time `gorm:"not null"`
}