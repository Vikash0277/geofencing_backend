package dto

type ConfigureAlertRequest struct {
	GeofenceID string  `json:"geofence_id"`
	VehicleID  *string `json:"vehicle_id"`
	EventType  string  `json:"event_type"`
	CreatedBy  string  `json:"created_by"`
}