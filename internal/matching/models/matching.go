package models

type MatchRequest struct {
	DriverID      string  `json:"driver_id"`
	OriginLat     float64 `json:"origin_lat"`
	OriginLng     float64 `json:"origin_lng"`
	DestLat       float64 `json:"dest_lat"`
	DestLng       float64 `json:"dest_lng"`
	RoutePolyline string  `json:"route_polyline"`
	MaxDetour     int     `json:"max_detour_percent"`
	Capacity      int     `json:"capacity"`
}

type PendingRide struct {
	RideID     string  `json:"ride_id"`
	RiderID    string  `json:"rider_id"`
	PickupLat  float64 `json:"pickup_lat"`
	PickupLng  float64 `json:"pickup_lng"`
	DropoffLat float64 `json:"dropoff_lat"`
	DropoffLng float64 `json:"dropoff_lng"`
	DetourCost float64 `json:"detour_cost_meters"`
}

type Batch struct {
	BatchID       string        `json:"batch_id"`
	DriverID      string        `json:"driver_id"`
	Riders        []PendingRide `json:"riders"`
	OptimizedRoute []Waypoint   `json:"optimized_route"`
	TotalDetourM  float64       `json:"total_detour_meters"`
	EstDurationS  int           `json:"estimated_duration_seconds"`
}

type Waypoint struct {
	Type    string  `json:"type"` // "pickup" or "dropoff"
	RideID  string  `json:"ride_id"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	ETA     int     `json:"eta_seconds"`
}
