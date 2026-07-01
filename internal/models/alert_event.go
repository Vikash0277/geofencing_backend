package models

import "time"

type AlertEvent struct {
	BaseModel

	AlertConfigID string    `gorm:"type:uuid;not null;index"`
	GeofenceID    string    `gorm:"type:uuid;not null;index"`
	VehicleID     string    `gorm:"type:uuid;not null;index"`
	UserID        string    `gorm:"type:uuid;not null;index"`
	EventType     string    `gorm:"type:varchar(50)"`
	Message       string    `gorm:"type:text"`
	Latitude      float64   `gorm:"not null"`
	Longitude     float64   `gorm:"not null"`
	Timestamp     time.Time `gorm:"not null"`
}