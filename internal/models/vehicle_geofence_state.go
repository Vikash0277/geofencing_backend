package models


type VehicleGeofenceState struct {
	BaseModel

	VehicleID string `gorm:"type:uuid;not null;index"`
	GeofenceID string `gorm:"type:uuid;not null;index"`

	IsInside bool `gorm:"not null"`
}