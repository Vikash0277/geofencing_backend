package models

import (
	"time"

	"github.com/google/uuid"
)

type BaseModel struct {
	ID          uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	CreatedAt   time.Time `gorm:"autoCreateTime;not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime;not null" json:"updated_at"`
}