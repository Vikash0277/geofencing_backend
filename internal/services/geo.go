package services

import (
	"fmt"
	"strings"

	"github.com/twpayne/go-geom"
)

// BuildWKT creates a WKT polygon string from coordinate pairs.
func BuildWKT(coords [][2]float64) string {
    var sb strings.Builder
    sb.WriteString("POLYGON((")
    for i, c := range coords {
        lng := c[1]
        lat := c[0]
        sb.WriteString(fmt.Sprintf("%f %f", lng, lat))
        if i != len(coords)-1 {
            sb.WriteString(", ")
        }
    }
    sb.WriteString("))")
    return sb.String()
}


func BuildPolygon(coords [][2]float64) (*geom.Polygon, error) {
    points := make([]geom.Coord, 0, len(coords))
    for _, c := range coords {
        lat := c[0]
        lng := c[1]
        points = append(points, geom.Coord{lng, lat})
    }
    poly := geom.NewPolygon(geom.XY)
    if p, err := poly.SetCoords([][]geom.Coord{points}); err != nil {
        return nil, err
    } else {
        p.SetSRID(4326)
        return p, nil
    }
}


func WKTToCoords(wkt string) [][2]float64 {

	// POLYGON((lng lat, lng lat, ...))
	wkt = strings.TrimPrefix(wkt, "POLYGON((")
	wkt = strings.TrimSuffix(wkt, "))")

	points := strings.Split(wkt, ",")

	result := make([][2]float64, 0, len(points))

	for _, p := range points {

		var lng, lat float64
		fmt.Sscanf(strings.TrimSpace(p), "%f %f", &lng, &lat)

		// return API format: [lat, lng]
		result = append(result, [2]float64{lat, lng})
	}

	return result
}

