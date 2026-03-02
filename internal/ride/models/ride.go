package models

import "time"

type RideStatus string

const (
	RideStatusRequested    RideStatus = "requested"
	RideStatusMatching     RideStatus = "matching"
	RideStatusMatched      RideStatus = "matched"
	RideStatusDriverEnRoute RideStatus = "driver_en_route"
	RideStatusPickupArrived RideStatus = "pickup_arrived"
	RideStatusInProgress   RideStatus = "in_progress"
	RideStatusCompleted    RideStatus = "completed"
	RideStatusCancelled    RideStatus = "cancelled"
	RideStatusDisputed     RideStatus = "disputed"
)

type Ride struct {
	ID          string     `json:"id"`
	RiderID     string     `json:"rider_id"`
	BatchID     *string    `json:"batch_id,omitempty"`
	Status      RideStatus `json:"status"`
	PickupLat   float64    `json:"pickup_lat"`
	PickupLng   float64    `json:"pickup_lng"`
	PickupAddr  string     `json:"pickup_addr"`
	DropoffLat  float64    `json:"dropoff_lat"`
	DropoffLng  float64    `json:"dropoff_lng"`
	DropoffAddr string     `json:"dropoff_addr"`
	FareAmount  float64    `json:"fare_amount"`
	IsRoundTrip bool       `json:"is_round_trip"`
	CreatedAt   time.Time  `json:"created_at"`
	MatchedAt   *time.Time `json:"matched_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type RideBatch struct {
	ID            string `json:"id"`
	DriverID      string `json:"driver_id"`
	Status        string `json:"status"`
	RoutePolyline string `json:"route_polyline,omitempty"`
	MaxPassengers int    `json:"max_passengers"`
	CurrentCount  int    `json:"current_count"`
	CreatedAt     time.Time `json:"created_at"`
}

type RideRequestInput struct {
	PickupLat   float64 `json:"pickup_lat" binding:"required"`
	PickupLng   float64 `json:"pickup_lng" binding:"required"`
	PickupAddr  string  `json:"pickup_addr"`
	DropoffLat  float64 `json:"dropoff_lat" binding:"required"`
	DropoffLng  float64 `json:"dropoff_lng" binding:"required"`
	DropoffAddr string  `json:"dropoff_addr"`
	IsRoundTrip bool    `json:"is_round_trip"`
}

type RideRating struct {
	ID         string `json:"id"`
	RideID     string `json:"ride_id"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Score      int    `json:"score" binding:"required,min=1,max=5"`
	Comment    string `json:"comment"`
}

type FareEstimate struct {
	BaseFare   float64 `json:"base_fare"`
	Surcharges float64 `json:"surcharges"`
	Discounts  float64 `json:"discounts"`
	Total      float64 `json:"total"`
	IsRoundTrip bool   `json:"is_round_trip"`
}
