package models

type User struct {
	BaseModel
	Name       string `gorm:"not null;type:varchar(50)" json:"name"`
	Email      string `gorm:"not null;unique;type:varchar(100)" json:"email"`
	Password   string `gorm:"type:varchar(255)" json:"password,omitempty"`
	Role       string `gorm:"default:'user';not null;type:varchar(20)" json:"role"`
	Provider   string `gorm:"default:'local';type:varchar(20)" json:"provider"`
	ProviderID string `gorm:"type:varchar(100)" json:"provider_id,omitempty"`
}
