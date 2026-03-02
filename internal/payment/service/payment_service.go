package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mohit1157/lama-yatayat-backend/internal/payment/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/payment/repository"
	"github.com/mohit1157/lama-yatayat-backend/pkg/events"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/refund"
)

type PaymentService struct {
	repo          *repository.PaymentRepository
	bus           events.Bus
	stripeEnabled bool
	commission    float64 // platform commission percentage (e.g., 0.08 for 8%)
}

func NewPaymentService(repo *repository.PaymentRepository, bus events.Bus, stripeKey string, commission float64) *PaymentService {
	svc := &PaymentService{
		repo:       repo,
		bus:        bus,
		commission: commission,
	}
	if stripeKey != "" && stripeKey != "your_stripe_secret_key_here" {
		stripe.Key = stripeKey
		svc.stripeEnabled = true
	}
	return svc
}

func (s *PaymentService) ChargeRider(ctx context.Context, req *models.ChargeRequest) (*models.Transaction, error) {
	txnID := uuid.New().String()
	txn := &models.Transaction{
		ID:     txnID,
		RideID: req.RideID,
		UserID: req.UserID,
		Type:   "charge",
		Amount: req.Amount,
		Status: "pending",
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	if s.stripeEnabled {
		// Create Stripe PaymentIntent
		pi, err := paymentintent.New(&stripe.PaymentIntentParams{
			Amount:   stripe.Int64(int64(req.Amount * 100)), // cents
			Currency: stripe.String("usd"),
			Metadata: map[string]string{
				"ride_id": req.RideID,
				"user_id": req.UserID,
				"txn_id":  txnID,
			},
			AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
				Enabled: stripe.Bool(true),
			},
		})
		if err != nil {
			s.repo.UpdateTransactionStatus(ctx, txnID, "failed", "")
			s.bus.Publish(ctx, "payment.failed", map[string]string{
				"ride_id": req.RideID, "user_id": req.UserID, "error": err.Error(),
			})
			return nil, fmt.Errorf("stripe charge failed: %w", err)
		}

		s.repo.UpdateTransactionStatus(ctx, txnID, "completed", pi.ID)
		txn.StripeID = pi.ID
		txn.Status = "completed"
	} else {
		// Demo mode: auto-complete
		s.repo.UpdateTransactionStatus(ctx, txnID, "completed", "demo_"+txnID)
		txn.Status = "completed"
		txn.StripeID = "demo_" + txnID
	}

	s.bus.Publish(ctx, "payment.completed", map[string]interface{}{
		"ride_id": req.RideID, "user_id": req.UserID, "amount": req.Amount, "txn_id": txnID,
	})

	return txn, nil
}

func (s *PaymentService) Refund(ctx context.Context, req *models.RefundRequest) (*models.Transaction, error) {
	original, err := s.repo.GetTransaction(ctx, req.TransactionID)
	if err != nil {
		return nil, err
	}
	if original.Status != "completed" {
		return nil, errors.New("can only refund completed transactions")
	}

	amount := req.Amount
	if amount <= 0 {
		amount = original.Amount
	}

	txnID := uuid.New().String()
	txn := &models.Transaction{
		ID:     txnID,
		RideID: original.RideID,
		UserID: original.UserID,
		Type:   "refund",
		Amount: amount,
		Status: "pending",
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, err
	}

	if s.stripeEnabled && original.StripeID != "" {
		r, err := refund.New(&stripe.RefundParams{
			PaymentIntent: stripe.String(original.StripeID),
			Amount:        stripe.Int64(int64(amount * 100)),
		})
		if err != nil {
			s.repo.UpdateTransactionStatus(ctx, txnID, "failed", "")
			return nil, fmt.Errorf("stripe refund failed: %w", err)
		}
		s.repo.UpdateTransactionStatus(ctx, txnID, "completed", r.ID)
		txn.StripeID = r.ID
	} else {
		s.repo.UpdateTransactionStatus(ctx, txnID, "completed", "demo_refund_"+txnID)
	}

	// Mark original as refunded
	s.repo.UpdateTransactionStatus(ctx, original.ID, "refunded", original.StripeID)
	txn.Status = "completed"

	return txn, nil
}

func (s *PaymentService) GetWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	return s.repo.GetWallet(ctx, userID)
}

func (s *PaymentService) PayoutDriver(ctx context.Context, req *models.PayoutRequest) (*models.Transaction, error) {
	txnID := uuid.New().String()
	txn := &models.Transaction{
		ID:     txnID,
		UserID: req.DriverID,
		Type:   "payout",
		Amount: req.Amount,
		Status: "completed", // Demo: instant payout
	}

	if err := s.repo.CreateTransaction(ctx, txn); err != nil {
		return nil, err
	}

	// Debit driver wallet
	if err := s.repo.DebitWallet(ctx, req.DriverID, req.Amount); err != nil {
		s.repo.UpdateTransactionStatus(ctx, txnID, "failed", "")
		return nil, err
	}

	return txn, nil
}

func (s *PaymentService) CreditDriverEarnings(ctx context.Context, driverID string, fareAmount float64) error {
	driverShare := fareAmount * (1 - s.commission)
	return s.repo.CreditWallet(ctx, driverID, driverShare)
}

func (s *PaymentService) GetHistory(ctx context.Context, userID string, limit, offset int) ([]models.Transaction, int, error) {
	return s.repo.GetHistory(ctx, userID, limit, offset)
}

// Payment methods

func (s *PaymentService) AddPaymentMethod(ctx context.Context, userID string, req *models.AddPaymentMethodRequest) (*models.PaymentMethod, error) {
	pm := &models.PaymentMethod{
		ID:        uuid.New().String(),
		UserID:    userID,
		Type:      "card",
		StripePM:  req.Token,
		LastFour:  "4242", // In production, extract from Stripe response
		IsDefault: true,
	}

	if err := s.repo.AddPaymentMethod(ctx, pm); err != nil {
		return nil, err
	}
	return pm, nil
}

func (s *PaymentService) ListPaymentMethods(ctx context.Context, userID string) ([]models.PaymentMethod, error) {
	return s.repo.ListPaymentMethods(ctx, userID)
}
