package dto

type CreateGeofenceRequest struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	Coordinates [][2]float64 `json:"coordinates"`
	CreatedBy   string       `json:"created_by"`
}

type GeofenceResponse struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Category    string       `json:"category"`
	Coordinates [][2]float64 `json:"coordinates"`
	CreatedAt   string       `json:"created_at"`
}