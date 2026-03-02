package models

type DriverLocation struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Heading   float64 `json:"heading"`
	Speed     float64 `json:"speed"`
	Timestamp int64   `json:"timestamp"`
}

type NearbyDriver struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance_meters"`
}

type RouteRequest struct {
	OriginLat  float64 `json:"origin_lat" binding:"required"`
	OriginLng  float64 `json:"origin_lng" binding:"required"`
	DestLat    float64 `json:"dest_lat" binding:"required"`
	DestLng    float64 `json:"dest_lng" binding:"required"`
}

type RouteResponse struct {
	Polyline     string  `json:"polyline"`
	DistanceM    float64 `json:"distance_meters"`
	DurationSec  int     `json:"duration_seconds"`
}

type ETAResponse struct {
	DurationSec int     `json:"duration_seconds"`
	DistanceM   float64 `json:"distance_meters"`
}
