package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Status       string    `json:"status"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type DriverProfile struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	LicenseNumber string     `json:"license_number"`
	LicenseDocURL string     `json:"license_doc_url,omitempty"`
	VehicleMake   string     `json:"vehicle_make"`
	VehicleModel  string     `json:"vehicle_model"`
	VehicleYear   int        `json:"vehicle_year"`
	VehiclePlate  string     `json:"vehicle_plate"`
	VehicleColor  string     `json:"vehicle_color"`
	Capacity      int        `json:"capacity"`
	BGCheckStatus string     `json:"bg_check_status"`
	RatingAvg     float64    `json:"rating_avg"`
	RatingCount   int        `json:"rating_count"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required,min=2"`
	Phone    string `json:"phone"`
	Role     string `json:"role" binding:"required,oneof=rider driver"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type OnboardDriverRequest struct {
	LicenseNumber string `json:"license_number" binding:"required"`
	VehicleMake   string `json:"vehicle_make" binding:"required"`
	VehicleModel  string `json:"vehicle_model" binding:"required"`
	VehicleYear   int    `json:"vehicle_year" binding:"required"`
	VehiclePlate  string `json:"vehicle_plate" binding:"required"`
	VehicleColor  string `json:"vehicle_color" binding:"required"`
	Capacity      int    `json:"capacity" binding:"required,min=1,max=6"`
}

type UpdateUserRequest struct {
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
}
