package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, email, phone, password_hash, name, role, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.Phone, user.PasswordHash, user.Name, user.Role, user.Status)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, COALESCE(phone,''), password_hash, name, role, status, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Name,
		&user.Role, &user.Status, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, COALESCE(phone,''), password_hash, name, role, status, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE email = $1`
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.Name,
		&user.Role, &user.Status, &user.AvatarURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, id string, req *models.UpdateUserRequest) error {
	query := `UPDATE users SET name = COALESCE(NULLIF($2, ''), name),
		phone = COALESCE(NULLIF($3, ''), phone),
		avatar_url = COALESCE(NULLIF($4, ''), avatar_url),
		updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, req.Name, req.Phone, req.AvatarURL)
	return err
}

func (r *UserRepository) ListByRole(ctx context.Context, role string, limit, offset int) ([]models.User, int, error) {
	var total int
	countQ := `SELECT COUNT(*) FROM users WHERE role = $1`
	r.db.QueryRow(ctx, countQ, role).Scan(&total)

	query := `SELECT id, email, COALESCE(phone,''), '', name, role, status, COALESCE(avatar_url,''), created_at, updated_at
		FROM users WHERE role = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, role, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Email, &u.Phone, &u.PasswordHash, &u.Name, &u.Role, &u.Status, &u.AvatarURL, &u.CreatedAt, &u.UpdatedAt)
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepository) CreateDriverProfile(ctx context.Context, dp *models.DriverProfile) error {
	query := `INSERT INTO driver_profiles (id, user_id, license_number, vehicle_make, vehicle_model,
		vehicle_year, vehicle_plate, vehicle_color, capacity, bg_check_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(ctx, query, dp.ID, dp.UserID, dp.LicenseNumber,
		dp.VehicleMake, dp.VehicleModel, dp.VehicleYear, dp.VehiclePlate, dp.VehicleColor, dp.Capacity, dp.BGCheckStatus)
	return err
}

func (r *UserRepository) GetDriverProfile(ctx context.Context, userID string) (*models.DriverProfile, error) {
	dp := &models.DriverProfile{}
	query := `SELECT id, user_id, license_number, vehicle_make, vehicle_model, vehicle_year,
		vehicle_plate, vehicle_color, capacity, bg_check_status, rating_avg, rating_count, verified_at, created_at
		FROM driver_profiles WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&dp.ID, &dp.UserID, &dp.LicenseNumber, &dp.VehicleMake, &dp.VehicleModel,
		&dp.VehicleYear, &dp.VehiclePlate, &dp.VehicleColor, &dp.Capacity,
		&dp.BGCheckStatus, &dp.RatingAvg, &dp.RatingCount, &dp.VerifiedAt, &dp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return dp, nil
}

func (r *UserRepository) UpdateDriverStatus(ctx context.Context, userID, status string) error {
	query := `UPDATE driver_profiles SET bg_check_status = $2, verified_at = CASE WHEN $2 = 'approved' THEN NOW() ELSE verified_at END WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID, status)
	return err
}
