package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/mohit1157/lama-yatayat-backend/internal/ride/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/ride/repository"
	"github.com/mohit1157/lama-yatayat-backend/pkg/events"
)

var (
	ErrRideNotFound    = errors.New("ride not found")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrCancelWindow    = errors.New("cancellation window expired")
)

const cancelGracePeriod = 2 * time.Minute

type RideService struct {
	repo *repository.RideRepository
	bus  events.Bus
}

func NewRideService(repo *repository.RideRepository, bus events.Bus) *RideService {
	return &RideService{repo: repo, bus: bus}
}

func (s *RideService) RequestRide(ctx context.Context, riderID string, req *models.RideRequestInput, fareAmount float64) (*models.Ride, error) {
	ride := &models.Ride{
		ID:          uuid.New().String(),
		RiderID:     riderID,
		Status:      models.RideStatusRequested,
		PickupLat:   req.PickupLat,
		PickupLng:   req.PickupLng,
		PickupAddr:  req.PickupAddr,
		DropoffLat:  req.DropoffLat,
		DropoffLng:  req.DropoffLng,
		DropoffAddr: req.DropoffAddr,
		FareAmount:  fareAmount,
		IsRoundTrip: req.IsRoundTrip,
	}

	if err := s.repo.Create(ctx, ride); err != nil {
		return nil, err
	}

	s.bus.Publish(ctx, "ride.requested", map[string]interface{}{
		"ride_id":    ride.ID,
		"rider_id":   riderID,
		"pickup_lat": req.PickupLat,
		"pickup_lng": req.PickupLng,
		"dropoff_lat": req.DropoffLat,
		"dropoff_lng": req.DropoffLng,
	})

	return ride, nil
}

func (s *RideService) GetRide(ctx context.Context, id string) (*models.Ride, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *RideService) CancelRide(ctx context.Context, id, userID string) error {
	ride, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrRideNotFound
	}

	// Only the rider or an admin can cancel
	if ride.RiderID != userID {
		return errors.New("not authorized to cancel this ride")
	}

	// Check if ride is in a cancellable state
	switch ride.Status {
	case models.RideStatusRequested, models.RideStatusMatching, models.RideStatusMatched, models.RideStatusDriverEnRoute:
		// Allowed
	default:
		return ErrInvalidTransition
	}

	// Check grace period for penalty-free cancellation
	if time.Since(ride.CreatedAt) > cancelGracePeriod &&
		(ride.Status == models.RideStatusDriverEnRoute || ride.Status == models.RideStatusMatched) {
		// Past grace period — still cancel but note penalty applies
		s.bus.Publish(ctx, "ride.cancelled.late", map[string]string{
			"ride_id": id, "rider_id": userID,
		})
	}

	if err := s.repo.UpdateStatus(ctx, id, models.RideStatusCancelled); err != nil {
		return err
	}

	s.bus.Publish(ctx, "ride.cancelled", map[string]string{
		"ride_id": id, "rider_id": userID,
	})
	return nil
}

func (s *RideService) ConfirmPickup(ctx context.Context, rideID string) error {
	ride, err := s.repo.GetByID(ctx, rideID)
	if err != nil {
		return ErrRideNotFound
	}

	if ride.Status != models.RideStatusPickupArrived && ride.Status != models.RideStatusDriverEnRoute {
		return ErrInvalidTransition
	}

	if err := s.repo.UpdateStatus(ctx, rideID, models.RideStatusInProgress); err != nil {
		return err
	}

	s.bus.Publish(ctx, "ride.started", map[string]string{
		"ride_id": rideID, "rider_id": ride.RiderID,
	})
	return nil
}

func (s *RideService) ConfirmDropoff(ctx context.Context, rideID string) error {
	ride, err := s.repo.GetByID(ctx, rideID)
	if err != nil {
		return ErrRideNotFound
	}

	if ride.Status != models.RideStatusInProgress {
		return ErrInvalidTransition
	}

	if err := s.repo.UpdateStatus(ctx, rideID, models.RideStatusCompleted); err != nil {
		return err
	}

	s.bus.Publish(ctx, "ride.completed", map[string]interface{}{
		"ride_id":     rideID,
		"rider_id":    ride.RiderID,
		"fare_amount": ride.FareAmount,
	})
	return nil
}

func (s *RideService) UpdateStatus(ctx context.Context, rideID string, status models.RideStatus) error {
	return s.repo.UpdateStatus(ctx, rideID, status)
}

func (s *RideService) GetActiveRide(ctx context.Context, riderID string) (*models.Ride, error) {
	return s.repo.GetActiveByRider(ctx, riderID)
}

func (s *RideService) GetHistory(ctx context.Context, riderID string, limit, offset int) ([]models.Ride, int, error) {
	return s.repo.GetHistory(ctx, riderID, limit, offset)
}

func (s *RideService) RateRide(ctx context.Context, rideID, fromUserID string, input *models.RateInput) error {
	ride, err := s.repo.GetByID(ctx, rideID)
	if err != nil {
		return ErrRideNotFound
	}

	if ride.Status != models.RideStatusCompleted {
		return errors.New("can only rate completed rides")
	}

	// Determine who is being rated
	driverID, err := s.repo.GetDriverForRide(ctx, rideID)
	if err != nil {
		return errors.New("driver not found for this ride")
	}

	rating := &models.RideRating{
		ID:         uuid.New().String(),
		RideID:     rideID,
		FromUserID: fromUserID,
		ToUserID:   driverID,
		Score:      input.Score,
		Comment:    input.Comment,
	}

	if err := s.repo.CreateRating(ctx, rating); err != nil {
		return err
	}

	// Update driver's aggregate rating
	return s.repo.UpdateDriverRating(ctx, driverID)
}

func (s *RideService) GetFareEstimate(ctx context.Context, isRoundTrip bool, baseFareRT, baseFareOW float64) *models.FareEstimate {
	base := baseFareOW
	if isRoundTrip {
		base = baseFareRT
	}
	return &models.FareEstimate{
		BaseFare:    base,
		Total:       base,
		IsRoundTrip: isRoundTrip,
	}
}

// Admin methods

func (s *RideService) ListAll(ctx context.Context, status string, limit, offset int) ([]models.Ride, int, error) {
	return s.repo.ListAll(ctx, status, limit, offset)
}

func (s *RideService) GetStats(ctx context.Context) (*models.RideStats, error) {
	return s.repo.GetStats(ctx)
}

// Batch operations

func (s *RideService) CreateBatch(ctx context.Context, batch *models.RideBatch) error {
	return s.repo.CreateBatch(ctx, batch)
}

func (s *RideService) GetBatch(ctx context.Context, id string) (*models.RideBatch, error) {
	return s.repo.GetBatch(ctx, id)
}

func (s *RideService) AssignRideToBatch(ctx context.Context, rideID, batchID string) error {
	if err := s.repo.SetBatch(ctx, rideID, batchID); err != nil {
		return err
	}
	return s.repo.IncrementBatchCount(ctx, batchID)
}
