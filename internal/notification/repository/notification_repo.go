package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mohit1157/lama-yatayat-backend/internal/notification/models"
)

type NotificationRepository struct {
	db *pgxpool.Pool
}

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, n *models.Notification) error {
	query := `INSERT INTO notifications (id, user_id, type, title, body, read)
		VALUES ($1, $2, $3, $4, $5, false)`
	_, err := r.db.Exec(ctx, query, n.ID, n.UserID, n.Type, n.Title, n.Body)
	return err
}

func (r *NotificationRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]models.Notification, int, error) {
	var total int
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id = $1`, userID).Scan(&total)

	query := `SELECT id, user_id, type, title, body, read, created_at
		FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var notifications []models.Notification
	for rows.Next() {
		var n models.Notification
		rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.Read, &n.CreatedAt)
		notifications = append(notifications, n)
	}
	return notifications, total, nil
}

func (r *NotificationRepository) MarkRead(ctx context.Context, id, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET read = true WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

// Push tokens

func (r *NotificationRepository) UpsertPushToken(ctx context.Context, pt *models.PushToken) error {
	query := `INSERT INTO push_tokens (id, user_id, token, platform)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, platform) DO UPDATE SET token = $3`
	_, err := r.db.Exec(ctx, query, pt.ID, pt.UserID, pt.Token, pt.Platform)
	return err
}

func (r *NotificationRepository) GetPushTokens(ctx context.Context, userID string) ([]models.PushToken, error) {
	query := `SELECT id, user_id, token, platform FROM push_tokens WHERE user_id = $1`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []models.PushToken
	for rows.Next() {
		var pt models.PushToken
		rows.Scan(&pt.ID, &pt.UserID, &pt.Token, &pt.Platform)
		tokens = append(tokens, pt)
	}
	return tokens, nil
}
