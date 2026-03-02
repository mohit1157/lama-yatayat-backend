package models

import "time"

type Transaction struct {
	ID        string    `json:"id"`
	RideID    string    `json:"ride_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // charge, refund, payout
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	StripeID  string    `json:"stripe_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ChargeRequest struct {
	RideID string  `json:"ride_id" binding:"required"`
	UserID string  `json:"user_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type Wallet struct {
	UserID         string  `json:"user_id"`
	Balance        float64 `json:"balance"`
	PendingBalance float64 `json:"pending_balance"`
}
