package services

import "fmt"

func BuildPointWKT(lat, lng float64) string {
	return fmt.Sprintf(
		"SRID=4326;POINT(%f %f)",
		lng,
		lat,
	)
}