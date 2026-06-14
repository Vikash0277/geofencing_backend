package models

type TrackMe struct {
	BaseModel
	VehicleNumber string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"vehicle_number"`
	DriverName    string    `gorm:"type:varchar(100);not null" json:"driver_name"`
	VehicleType   string    `gorm:"type:varchar(50);not null" json:"vehicle_type"`
	Phone         string    `gorm:"type:varchar(15);not null" json:"phone"`
	Status        string    `gorm:"type:varchar(20);not null;default:active" json:"status"`
}
