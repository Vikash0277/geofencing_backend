package services

import "fmt"

func ValidateCoordinates(coords [][2]float64) error {

	if len(coords) < 4 {
		return fmt.Errorf("minimum 4 coordinates required")
	}

	first := coords[0]
	last := coords[len(coords)-1]

	if first != last {
		return fmt.Errorf("polygon must be closed (first and last point must match)")
	}

	for _, c := range coords {
		lat := c[0]
		lng := c[1]

		if lat < -90 || lat > 90 {
			return fmt.Errorf("invalid latitude")
		}

		if lng < -180 || lng > 180 {
			return fmt.Errorf("invalid longitude")
		}
	}

	return nil
}