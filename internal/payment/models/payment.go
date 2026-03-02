package models

import "time"

type Transaction struct {
	ID        string    `json:"id"`
	RideID    string    `json:"ride_id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"` // charge, refund, payout, credit
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // pending, completed, failed, refunded
	StripeID  string    `json:"stripe_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ChargeRequest struct {
	RideID string  `json:"ride_id" binding:"required"`
	UserID string  `json:"user_id" binding:"required"`
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type RefundRequest struct {
	TransactionID string  `json:"transaction_id" binding:"required"`
	Amount        float64 `json:"amount,omitempty"` // 0 = full refund
}

type Wallet struct {
	UserID         string  `json:"user_id"`
	Balance        float64 `json:"balance"`
	PendingBalance float64 `json:"pending_balance"`
	Currency       string  `json:"currency"`
}

type PaymentMethod struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	LastFour  string    `json:"last_four"`
	StripePM  string    `json:"stripe_pm,omitempty"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at"`
}

type AddPaymentMethodRequest struct {
	Token string `json:"token" binding:"required"`
}

type PayoutRequest struct {
	DriverID string  `json:"driver_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
}
