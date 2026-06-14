package models

import (
	"github.com/google/uuid"
)

type Geofence struct {
	BaseModel

	Name        string `gorm:"type:varchar(50)" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`

	Category string `gorm:"type:varchar(30);not null" json:"category"`
	Status   string `gorm:"type:varchar(20);default:'active';not null" json:"status"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
	
	Polygon string `gorm:"type:geometry(Polygon,4326);not null" json:"-"`
}