package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/user/repository"
	"github.com/mohit1157/lama-yatayat-backend/pkg/auth"
	"github.com/mohit1157/lama-yatayat-backend/pkg/events"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken     = errors.New("email already registered")
	ErrInvalidCreds   = errors.New("invalid email or password")
	ErrNotDriver      = errors.New("user is not a driver")
	ErrAlreadyOnboard = errors.New("driver already onboarded")
)

type UserService struct {
	repo   *repository.UserRepository
	jwt    *auth.JWTManager
	bus    events.Bus
}

func NewUserService(repo *repository.UserRepository, jwt *auth.JWTManager, bus events.Bus) *UserService {
	return &UserService{repo: repo, jwt: jwt, bus: bus}
}

func (s *UserService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Check if email exists
	existing, _ := s.repo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hash),
		Name:         req.Name,
		Role:         req.Role,
		Status:       "active",
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Generate tokens
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	// Publish event
	s.bus.Publish(ctx, "user.registered", map[string]string{
		"user_id": user.ID, "role": user.Role, "email": user.Email,
	})

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	accessToken, _ := s.jwt.GenerateAccessToken(user.ID, user.Role)
	refreshToken, _ := s.jwt.GenerateRefreshToken(user.ID, user.Role)

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req *models.UpdateUserRequest) error {
	return s.repo.Update(ctx, id, req)
}

func (s *UserService) OnboardDriver(ctx context.Context, userID string, req *models.OnboardDriverRequest) (*models.DriverProfile, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Role != "driver" {
		return nil, ErrNotDriver
	}

	existing, _ := s.repo.GetDriverProfile(ctx, userID)
	if existing != nil {
		return nil, ErrAlreadyOnboard
	}

	dp := &models.DriverProfile{
		ID:            uuid.New().String(),
		UserID:        userID,
		LicenseNumber: req.LicenseNumber,
		VehicleMake:   req.VehicleMake,
		VehicleModel:  req.VehicleModel,
		VehicleYear:   req.VehicleYear,
		VehiclePlate:  req.VehiclePlate,
		VehicleColor:  req.VehicleColor,
		Capacity:      req.Capacity,
		BGCheckStatus: "pending",
	}

	if err := s.repo.CreateDriverProfile(ctx, dp); err != nil {
		return nil, err
	}

	return dp, nil
}

func (s *UserService) GetDriverStatus(ctx context.Context, userID string) (*models.DriverProfile, error) {
	return s.repo.GetDriverProfile(ctx, userID)
}

func (s *UserService) ApproveDriver(ctx context.Context, userID string) error {
	if err := s.repo.UpdateDriverStatus(ctx, userID, "approved"); err != nil {
		return err
	}
	s.bus.Publish(ctx, "driver.verified", map[string]string{"user_id": userID})
	return nil
}

func (s *UserService) SuspendDriver(ctx context.Context, userID string) error {
	if err := s.repo.UpdateDriverStatus(ctx, userID, "suspended"); err != nil {
		return err
	}
	s.bus.Publish(ctx, "driver.suspended", map[string]string{"user_id": userID})
	return nil
}

func (s *UserService) ListDrivers(ctx context.Context, limit, offset int) ([]models.DriverWithUser, int, error) {
	return s.repo.ListDriversWithProfiles(ctx, limit, offset)
}

func (s *UserService) ListUsers(ctx context.Context, limit, offset int) ([]models.User, int, error) {
	return s.repo.ListByRole(ctx, "rider", limit, offset)
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*models.AuthResponse, error) {
	claims, err := s.jwt.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	newAccess, _ := s.jwt.GenerateAccessToken(user.ID, user.Role)
	newRefresh, _ := s.jwt.GenerateRefreshToken(user.ID, user.Role)

	return &models.AuthResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
		User:         *user,
	}, nil
}
