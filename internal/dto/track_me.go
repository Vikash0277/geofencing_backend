package dto

type CreateTrackMeDTO struct {
	VehicleNumber string `json:"vehicle_number" validate:"required"`
	DriverName    string `json:"driver_name" validate:"required"`
	VehicleType   string `json:"vehicle_type" validate:"required"`
	Phone         string `json:"phone" validate:"required"`
	Status        string `json:"status"`
}

type UpdateTrackMeDTO struct {
	DriverName  string `json:"driver_name"`
	VehicleType string `json:"vehicle_type"`
	Phone       string `json:"phone"`
	Status      string `json:"status"`
}