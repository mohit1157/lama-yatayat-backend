package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohit1157/lama-yatayat-backend/internal/payment/models"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) CreateTransaction(ctx context.Context, txn *models.Transaction) error {
	query := `INSERT INTO transactions (id, ride_id, user_id, type, amount, status, stripe_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.Exec(ctx, query, txn.ID, txn.RideID, txn.UserID, txn.Type, txn.Amount, txn.Status, txn.StripeID)
	return err
}

func (r *PaymentRepository) GetTransaction(ctx context.Context, id string) (*models.Transaction, error) {
	txn := &models.Transaction{}
	query := `SELECT id, ride_id, user_id, type, amount, status, stripe_id, created_at
		FROM transactions WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&txn.ID, &txn.RideID, &txn.UserID, &txn.Type, &txn.Amount, &txn.Status, &txn.StripeID, &txn.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}
	return txn, nil
}

func (r *PaymentRepository) UpdateTransactionStatus(ctx context.Context, id, status, stripeID string) error {
	query := `UPDATE transactions SET status = $2, stripe_id = COALESCE(NULLIF($3, ''), stripe_id) WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status, stripeID)
	return err
}

func (r *PaymentRepository) GetWallet(ctx context.Context, userID string) (*models.Wallet, error) {
	w := &models.Wallet{}
	query := `SELECT user_id, balance, pending_balance, currency FROM wallets WHERE user_id = $1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&w.UserID, &w.Balance, &w.PendingBalance, &w.Currency)
	if err != nil {
		// Create wallet if not exists
		_, createErr := r.db.Exec(ctx,
			`INSERT INTO wallets (id, user_id, balance, pending_balance, currency) VALUES (gen_random_uuid(), $1, 0, 0, 'USD') ON CONFLICT (user_id) DO NOTHING`, userID)
		if createErr != nil {
			return nil, createErr
		}
		return &models.Wallet{UserID: userID, Balance: 0, PendingBalance: 0, Currency: "USD"}, nil
	}
	return w, nil
}

func (r *PaymentRepository) CreditWallet(ctx context.Context, userID string, amount float64) error {
	query := `UPDATE wallets SET balance = balance + $2, updated_at = NOW() WHERE user_id = $1`
	tag, err := r.db.Exec(ctx, query, userID, amount)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		// Wallet doesn't exist, create it with this amount
		_, err = r.db.Exec(ctx,
			`INSERT INTO wallets (id, user_id, balance, pending_balance, currency) VALUES (gen_random_uuid(), $1, $2, 0, 'USD')`, userID, amount)
		return err
	}
	return nil
}

func (r *PaymentRepository) DebitWallet(ctx context.Context, userID string, amount float64) error {
	query := `UPDATE wallets SET balance = balance - $2, updated_at = NOW() WHERE user_id = $1 AND balance >= $2`
	tag, err := r.db.Exec(ctx, query, userID, amount)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient balance")
	}
	return nil
}

func (r *PaymentRepository) GetHistory(ctx context.Context, userID string, limit, offset int) ([]models.Transaction, int, error) {
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM transactions WHERE user_id = $1`, userID).Scan(&total)

	query := `SELECT id, ride_id, user_id, type, amount, status, stripe_id, created_at
		FROM transactions WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var txns []models.Transaction
	for rows.Next() {
		var t models.Transaction
		rows.Scan(&t.ID, &t.RideID, &t.UserID, &t.Type, &t.Amount, &t.Status, &t.StripeID, &t.CreatedAt)
		txns = append(txns, t)
	}
	return txns, total, nil
}

// Payment methods

func (r *PaymentRepository) AddPaymentMethod(ctx context.Context, pm *models.PaymentMethod) error {
	// If this is set as default, unset existing defaults
	if pm.IsDefault {
		r.db.Exec(ctx, `UPDATE payment_methods SET is_default = false WHERE user_id = $1`, pm.UserID)
	}
	query := `INSERT INTO payment_methods (id, user_id, type, stripe_pm, last_four, is_default)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, pm.ID, pm.UserID, pm.Type, pm.StripePM, pm.LastFour, pm.IsDefault)
	return err
}

func (r *PaymentRepository) ListPaymentMethods(ctx context.Context, userID string) ([]models.PaymentMethod, error) {
	query := `SELECT id, user_id, type, last_four, is_default, created_at
		FROM payment_methods WHERE user_id = $1 ORDER BY is_default DESC, created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var methods []models.PaymentMethod
	for rows.Next() {
		var pm models.PaymentMethod
		rows.Scan(&pm.ID, &pm.UserID, &pm.Type, &pm.LastFour, &pm.IsDefault, &pm.CreatedAt)
		methods = append(methods, pm)
	}
	return methods, nil
}

func (r *PaymentRepository) GetDefaultPaymentMethod(ctx context.Context, userID string) (*models.PaymentMethod, error) {
	pm := &models.PaymentMethod{}
	query := `SELECT id, user_id, type, stripe_pm, last_four, is_default, created_at
		FROM payment_methods WHERE user_id = $1 AND is_default = true LIMIT 1`
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&pm.ID, &pm.UserID, &pm.Type, &pm.StripePM, &pm.LastFour, &pm.IsDefault, &pm.CreatedAt)
	if err != nil {
		return nil, err
	}
	return pm, nil
}
