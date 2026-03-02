package engine

import (
	"math"

	"github.com/mohit1157/lama-yatayat-backend/internal/matching/models"
)

// Haversine returns distance in meters between two lat/lng points
func Haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// PointToSegmentDistance returns the perpendicular distance from a point
// to a line segment (in meters). Used for corridor matching.
func PointToSegmentDistance(pLat, pLng, aLat, aLng, bLat, bLng float64) float64 {
	dAP := Haversine(aLat, aLng, pLat, pLng)
	dAB := Haversine(aLat, aLng, bLat, bLng)
	dBP := Haversine(bLat, bLng, pLat, pLng)

	if dAB == 0 {
		return dAP
	}

	// Project point onto line segment
	t := ((pLat-aLat)*(bLat-aLat) + (pLng-aLng)*(bLng-aLng)) /
		((bLat-aLat)*(bLat-aLat) + (bLng-aLng)*(bLng-aLng))

	if t < 0 {
		return dAP
	}
	if t > 1 {
		return dBP
	}

	projLat := aLat + t*(bLat-aLat)
	projLng := aLng + t*(bLng-aLng)
	return Haversine(pLat, pLng, projLat, projLng)
}

// FilterByCorridorDistance filters pending rides that fall within
// corridorMeters of any segment in the route polyline.
func FilterByCorridorDistance(
	rides []models.PendingRide,
	routePoints [][2]float64, // [[lat, lng], ...]
	corridorMeters float64,
) []models.PendingRide {
	var result []models.PendingRide

	for _, ride := range rides {
		minDist := math.MaxFloat64

		for i := 0; i < len(routePoints)-1; i++ {
			d := PointToSegmentDistance(
				ride.PickupLat, ride.PickupLng,
				routePoints[i][0], routePoints[i][1],
				routePoints[i+1][0], routePoints[i+1][1],
			)
			if d < minDist {
				minDist = d
			}
		}

		if minDist <= corridorMeters {
			ride.DetourCost = minDist
			result = append(result, ride)
		}
	}

	return result
}
