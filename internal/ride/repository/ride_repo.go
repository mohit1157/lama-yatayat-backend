package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohit1157/lama-yatayat-backend/internal/ride/models"
)

type RideRepository struct {
	db *pgxpool.Pool
}

func NewRideRepository(db *pgxpool.Pool) *RideRepository {
	return &RideRepository{db: db}
}

func (r *RideRepository) Create(ctx context.Context, ride *models.Ride) error {
	query := `INSERT INTO rides (id, rider_id, status, pickup_lat, pickup_lng, pickup_addr,
		dropoff_lat, dropoff_lng, dropoff_addr, fare_amount, is_round_trip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := r.db.Exec(ctx, query, ride.ID, ride.RiderID, ride.Status,
		ride.PickupLat, ride.PickupLng, ride.PickupAddr,
		ride.DropoffLat, ride.DropoffLng, ride.DropoffAddr,
		ride.FareAmount, ride.IsRoundTrip)
	return err
}

func (r *RideRepository) GetByID(ctx context.Context, id string) (*models.Ride, error) {
	ride := &models.Ride{}
	query := `SELECT id, rider_id, batch_id, status, pickup_lat, pickup_lng, pickup_addr,
		dropoff_lat, dropoff_lng, dropoff_addr, fare_amount, is_round_trip,
		created_at, matched_at, started_at, completed_at, cancelled_at
		FROM rides WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ride.ID, &ride.RiderID, &ride.BatchID, &ride.Status,
		&ride.PickupLat, &ride.PickupLng, &ride.PickupAddr,
		&ride.DropoffLat, &ride.DropoffLng, &ride.DropoffAddr,
		&ride.FareAmount, &ride.IsRoundTrip,
		&ride.CreatedAt, &ride.MatchedAt, &ride.StartedAt, &ride.CompletedAt, &ride.CancelledAt)
	if err != nil {
		return nil, fmt.Errorf("ride not found: %w", err)
	}
	return ride, nil
}

func (r *RideRepository) UpdateStatus(ctx context.Context, id string, status models.RideStatus) error {
	var tsCol string
	switch status {
	case models.RideStatusMatched:
		tsCol = ", matched_at = NOW()"
	case models.RideStatusInProgress:
		tsCol = ", started_at = NOW()"
	case models.RideStatusCompleted:
		tsCol = ", completed_at = NOW()"
	case models.RideStatusCancelled:
		tsCol = ", cancelled_at = NOW()"
	default:
		tsCol = ""
	}
	query := fmt.Sprintf(`UPDATE rides SET status = $2%s WHERE id = $1`, tsCol)
	tag, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("ride not found")
	}
	return nil
}

func (r *RideRepository) SetBatch(ctx context.Context, rideID, batchID string) error {
	query := `UPDATE rides SET batch_id = $2, status = 'matched', matched_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, rideID, batchID)
	return err
}

func (r *RideRepository) GetActiveByRider(ctx context.Context, riderID string) (*models.Ride, error) {
	ride := &models.Ride{}
	query := `SELECT id, rider_id, batch_id, status, pickup_lat, pickup_lng, pickup_addr,
		dropoff_lat, dropoff_lng, dropoff_addr, fare_amount, is_round_trip,
		created_at, matched_at, started_at, completed_at, cancelled_at
		FROM rides WHERE rider_id = $1
		AND status NOT IN ('completed', 'cancelled', 'disputed')
		ORDER BY created_at DESC LIMIT 1`
	err := r.db.QueryRow(ctx, query, riderID).Scan(
		&ride.ID, &ride.RiderID, &ride.BatchID, &ride.Status,
		&ride.PickupLat, &ride.PickupLng, &ride.PickupAddr,
		&ride.DropoffLat, &ride.DropoffLng, &ride.DropoffAddr,
		&ride.FareAmount, &ride.IsRoundTrip,
		&ride.CreatedAt, &ride.MatchedAt, &ride.StartedAt, &ride.CompletedAt, &ride.CancelledAt)
	if err != nil {
		return nil, err
	}
	return ride, nil
}

func (r *RideRepository) GetHistory(ctx context.Context, riderID string, limit, offset int) ([]models.Ride, int, error) {
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rides WHERE rider_id = $1`, riderID).Scan(&total)

	query := `SELECT id, rider_id, batch_id, status, pickup_lat, pickup_lng, pickup_addr,
		dropoff_lat, dropoff_lng, dropoff_addr, fare_amount, is_round_trip,
		created_at, matched_at, started_at, completed_at, cancelled_at
		FROM rides WHERE rider_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, riderID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rides []models.Ride
	for rows.Next() {
		var rd models.Ride
		rows.Scan(&rd.ID, &rd.RiderID, &rd.BatchID, &rd.Status,
			&rd.PickupLat, &rd.PickupLng, &rd.PickupAddr,
			&rd.DropoffLat, &rd.DropoffLng, &rd.DropoffAddr,
			&rd.FareAmount, &rd.IsRoundTrip,
			&rd.CreatedAt, &rd.MatchedAt, &rd.StartedAt, &rd.CompletedAt, &rd.CancelledAt)
		rides = append(rides, rd)
	}
	return rides, total, nil
}

func (r *RideRepository) CreateRating(ctx context.Context, rating *models.RideRating) error {
	query := `INSERT INTO ride_ratings (id, ride_id, from_user_id, to_user_id, score, comment)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, rating.ID, rating.RideID, rating.FromUserID, rating.ToUserID, rating.Score, rating.Comment)
	return err
}

func (r *RideRepository) UpdateDriverRating(ctx context.Context, driverUserID string) error {
	query := `UPDATE driver_profiles SET
		rating_avg = (SELECT COALESCE(AVG(score), 0) FROM ride_ratings WHERE to_user_id = $1),
		rating_count = (SELECT COUNT(*) FROM ride_ratings WHERE to_user_id = $1)
		WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, driverUserID)
	return err
}

// GetDriverForRide returns the driver's user_id for a ride via its batch
func (r *RideRepository) GetDriverForRide(ctx context.Context, rideID string) (string, error) {
	var driverID string
	query := `SELECT rb.driver_id FROM ride_batches rb
		JOIN rides r ON r.batch_id = rb.id WHERE r.id = $1`
	err := r.db.QueryRow(ctx, query, rideID).Scan(&driverID)
	return driverID, err
}

// Admin queries

func (r *RideRepository) ListAll(ctx context.Context, status string, limit, offset int) ([]models.Ride, int, error) {
	var total int
	countQ := `SELECT COUNT(*) FROM rides WHERE ($1 = '' OR status = $1)`
	r.db.QueryRow(ctx, countQ, status).Scan(&total)

	query := `SELECT id, rider_id, batch_id, status, pickup_lat, pickup_lng, pickup_addr,
		dropoff_lat, dropoff_lng, dropoff_addr, fare_amount, is_round_trip,
		created_at, matched_at, started_at, completed_at, cancelled_at
		FROM rides WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var rides []models.Ride
	for rows.Next() {
		var rd models.Ride
		rows.Scan(&rd.ID, &rd.RiderID, &rd.BatchID, &rd.Status,
			&rd.PickupLat, &rd.PickupLng, &rd.PickupAddr,
			&rd.DropoffLat, &rd.DropoffLng, &rd.DropoffAddr,
			&rd.FareAmount, &rd.IsRoundTrip,
			&rd.CreatedAt, &rd.MatchedAt, &rd.StartedAt, &rd.CompletedAt, &rd.CancelledAt)
		rides = append(rides, rd)
	}
	return rides, total, nil
}

func (r *RideRepository) GetStats(ctx context.Context) (*models.RideStats, error) {
	stats := &models.RideStats{}
	today := time.Now().Truncate(24 * time.Hour)
	weekAgo := today.AddDate(0, 0, -7)

	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rides WHERE created_at >= $1`, today).Scan(&stats.RidesToday)
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rides WHERE created_at >= $1`, weekAgo).Scan(&stats.RidesWeek)
	r.db.QueryRow(ctx, `SELECT COALESCE(SUM(fare_amount), 0) FROM rides WHERE status = 'completed' AND completed_at >= $1`, today).Scan(&stats.RevenueToday)
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rides WHERE status NOT IN ('completed', 'cancelled', 'disputed')`).Scan(&stats.ActiveRides)
	r.db.QueryRow(ctx, `SELECT COALESCE(AVG(current_count), 0) FROM ride_batches WHERE status = 'completed'`).Scan(&stats.AvgPassengersPerBatch)

	return stats, nil
}

// Batch operations

func (r *RideRepository) CreateBatch(ctx context.Context, batch *models.RideBatch) error {
	query := `INSERT INTO ride_batches (id, driver_id, status, route_polyline, max_passengers, current_count)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, batch.ID, batch.DriverID, batch.Status,
		batch.RoutePolyline, batch.MaxPassengers, batch.CurrentCount)
	return err
}

func (r *RideRepository) GetBatch(ctx context.Context, id string) (*models.RideBatch, error) {
	batch := &models.RideBatch{}
	query := `SELECT id, driver_id, status, route_polyline, max_passengers, current_count, created_at
		FROM ride_batches WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&batch.ID, &batch.DriverID, &batch.Status, &batch.RoutePolyline,
		&batch.MaxPassengers, &batch.CurrentCount, &batch.CreatedAt)
	if err != nil {
		return nil, err
	}
	return batch, nil
}

func (r *RideRepository) UpdateBatchStatus(ctx context.Context, id, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE ride_batches SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *RideRepository) IncrementBatchCount(ctx context.Context, batchID string) error {
	_, err := r.db.Exec(ctx, `UPDATE ride_batches SET current_count = current_count + 1 WHERE id = $1`, batchID)
	return err
}
