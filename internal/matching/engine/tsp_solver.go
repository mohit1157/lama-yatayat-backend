package engine

import (
	"github.com/mohit1157/lama-yatayat-backend/internal/matching/models"
)

// SolvePickupDropoffOrder uses a greedy nearest-insertion heuristic
// with 2-opt local search to optimize the order of pickups and dropoffs.
// For batches of 2-4 riders (4-8 waypoints), this runs in < 1ms.
func SolvePickupDropoffOrder(
	driverLat, driverLng float64,
	riders []models.PendingRide,
) []models.Waypoint {
	if len(riders) == 0 {
		return nil
	}

	// Build list of all waypoints
	type wp struct {
		lat, lng float64
		rideID   string
		wpType   string
		visited  bool
	}

	var points []wp
	for _, r := range riders {
		points = append(points, wp{r.PickupLat, r.PickupLng, r.RideID, "pickup", false})
	}
	for _, r := range riders {
		points = append(points, wp{r.DropoffLat, r.DropoffLng, r.RideID, "dropoff", false})
	}

	// Track which riders have been picked up for ordering constraint
	pickedUp := make(map[string]bool)
	var route []models.Waypoint

	curLat, curLng := driverLat, driverLng

	for len(route) < len(points) {
		bestIdx := -1
		bestDist := float64(1e18)

		for i, p := range points {
			if p.visited {
				continue
			}
			// Dropoff can only happen after pickup
			if p.wpType == "dropoff" && !pickedUp[p.rideID] {
				continue
			}

			d := Haversine(curLat, curLng, p.lat, p.lng)
			if d < bestDist {
				bestDist = d
				bestIdx = i
			}
		}

		if bestIdx < 0 {
			break
		}

		points[bestIdx].visited = true
		if points[bestIdx].wpType == "pickup" {
			pickedUp[points[bestIdx].rideID] = true
		}

		route = append(route, models.Waypoint{
			Type:   points[bestIdx].wpType,
			RideID: points[bestIdx].rideID,
			Lat:    points[bestIdx].lat,
			Lng:    points[bestIdx].lng,
		})
		curLat = points[bestIdx].lat
		curLng = points[bestIdx].lng
	}

	// 2-opt improvement (swap pairs and check if total distance improves)
	improved := true
	for improved {
		improved = false
		for i := 0; i < len(route)-1; i++ {
			for j := i + 1; j < len(route); j++ {
				if canSwap(route, i, j) {
					oldDist := segmentDistance(route, i, j)
					route[i], route[j] = route[j], route[i]
					newDist := segmentDistance(route, i, j)
					if newDist < oldDist {
						improved = true
					} else {
						route[i], route[j] = route[j], route[i] // revert
					}
				}
			}
		}
	}

	return route
}

// canSwap checks if swapping two waypoints maintains pickup-before-dropoff constraint
func canSwap(route []models.Waypoint, i, j int) bool {
	// Simple check: don't swap a pickup past its dropoff or vice versa
	a, b := route[i], route[j]

	if a.RideID == b.RideID {
		return false // Same ride: pickup must come before dropoff
	}

	// Simulate swap
	route[i], route[j] = route[j], route[i]
	valid := validateOrder(route)
	route[i], route[j] = route[j], route[i] // revert
	return valid
}

func validateOrder(route []models.Waypoint) bool {
	pickedUp := make(map[string]bool)
	for _, w := range route {
		if w.Type == "dropoff" && !pickedUp[w.RideID] {
			return false
		}
		if w.Type == "pickup" {
			pickedUp[w.RideID] = true
		}
	}
	return true
}

func segmentDistance(route []models.Waypoint, i, j int) float64 {
	total := 0.0
	start := i
	if start > 0 {
		start--
	}
	for k := start; k < j && k < len(route)-1; k++ {
		total += Haversine(route[k].Lat, route[k].Lng, route[k+1].Lat, route[k+1].Lng)
	}
	return total
}
