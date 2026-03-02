package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/mohit1157/lama-yatayat-backend/internal/matching/engine"
	"github.com/mohit1157/lama-yatayat-backend/internal/matching/models"
	"github.com/mohit1157/lama-yatayat-backend/pkg/events"
)

const pendingRidesKey = "pending_rides"

type MatchingService struct {
	rdb            *redis.Client
	bus            events.Bus
	corridorMeters float64
	maxBatchSize   int
}

func NewMatchingService(rdb *redis.Client, bus events.Bus, corridorMeters float64, maxBatchSize int) *MatchingService {
	return &MatchingService{
		rdb:            rdb,
		bus:            bus,
		corridorMeters: corridorMeters,
		maxBatchSize:   maxBatchSize,
	}
}

// AddPendingRide stores a ride request in Redis for matching
func (s *MatchingService) AddPendingRide(ctx context.Context, ride models.PendingRide) error {
	data, _ := json.Marshal(ride)
	return s.rdb.HSet(ctx, pendingRidesKey, ride.RideID, data).Err()
}

// RemovePendingRide removes a matched or cancelled ride
func (s *MatchingService) RemovePendingRide(ctx context.Context, rideID string) error {
	return s.rdb.HDel(ctx, pendingRidesKey, rideID).Err()
}

// FindRiders matches pending rides to a driver's route
func (s *MatchingService) FindRiders(ctx context.Context, req *models.MatchRequest) (*models.Batch, error) {
	// Get all pending rides from Redis
	pendingData, err := s.rdb.HGetAll(ctx, pendingRidesKey).Result()
	if err != nil {
		return nil, fmt.Errorf("fetch pending rides: %w", err)
	}

	var pending []models.PendingRide
	for _, data := range pendingData {
		var ride models.PendingRide
		if err := json.Unmarshal([]byte(data), &ride); err == nil {
			pending = append(pending, ride)
		}
	}

	if len(pending) == 0 {
		return nil, fmt.Errorf("no pending rides")
	}

	// Build route points from polyline (for demo, use origin→destination as 2-point route)
	routePoints := [][2]float64{
		{req.OriginLat, req.OriginLng},
		{req.DestLat, req.DestLng},
	}

	// Filter by corridor distance
	filtered := engine.FilterByCorridorDistance(pending, routePoints, s.corridorMeters)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("no riders along route")
	}

	// Cap to batch size
	if len(filtered) > s.maxBatchSize {
		filtered = filtered[:s.maxBatchSize]
	}

	// Optimize pickup/dropoff order with TSP solver
	optimizedRoute := engine.SolvePickupDropoffOrder(req.OriginLat, req.OriginLng, filtered)

	// Create batch
	batch := &models.Batch{
		BatchID:        uuid.New().String(),
		DriverID:       req.DriverID,
		Riders:         filtered,
		OptimizedRoute: optimizedRoute,
	}

	// Calculate total detour
	for _, r := range filtered {
		batch.TotalDetourM += r.DetourCost
	}

	// Publish batch created event
	s.bus.Publish(ctx, "batch.created", map[string]interface{}{
		"batch_id":  batch.BatchID,
		"driver_id": req.DriverID,
		"riders":    len(filtered),
	})

	log.Printf("MATCH: Created batch %s with %d riders for driver %s", batch.BatchID, len(filtered), req.DriverID)

	return batch, nil
}

// SetupEventListeners subscribes to ride events for automatic matching
func (s *MatchingService) SetupEventListeners() {
	s.bus.Subscribe("ride.requested", func(ctx context.Context, event events.Event) error {
		var payload struct {
			RideID     string  `json:"ride_id"`
			RiderID    string  `json:"rider_id"`
			PickupLat  float64 `json:"pickup_lat"`
			PickupLng  float64 `json:"pickup_lng"`
			DropoffLat float64 `json:"dropoff_lat"`
			DropoffLng float64 `json:"dropoff_lng"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return err
		}

		return s.AddPendingRide(ctx, models.PendingRide{
			RideID:     payload.RideID,
			RiderID:    payload.RiderID,
			PickupLat:  payload.PickupLat,
			PickupLng:  payload.PickupLng,
			DropoffLat: payload.DropoffLat,
			DropoffLng: payload.DropoffLng,
		})
	})

	s.bus.Subscribe("ride.cancelled", func(ctx context.Context, event events.Event) error {
		var payload struct {
			RideID string `json:"ride_id"`
		}
		json.Unmarshal(event.Payload, &payload)
		return s.RemovePendingRide(ctx, payload.RideID)
	})
}
