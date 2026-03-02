package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/mohit1157/lama-yatayat-backend/internal/notification/models"
	"github.com/mohit1157/lama-yatayat-backend/internal/notification/repository"
)

const expoPushURL = "https://exp.host/--/api/v2/push/send"

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) RegisterToken(ctx context.Context, userID string, req *models.RegisterTokenRequest) error {
	pt := &models.PushToken{
		ID:       uuid.New().String(),
		UserID:   userID,
		Token:    req.Token,
		Platform: req.Platform,
	}
	return s.repo.UpsertPushToken(ctx, pt)
}

func (s *NotificationService) ListNotifications(ctx context.Context, userID string, limit, offset int) ([]models.Notification, int, error) {
	return s.repo.ListByUser(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkRead(ctx context.Context, id, userID string) error {
	return s.repo.MarkRead(ctx, id, userID)
}

func (s *NotificationService) SendPush(ctx context.Context, req *models.SendPushRequest) error {
	// Store in DB
	notif := &models.Notification{
		ID:     uuid.New().String(),
		UserID: req.UserID,
		Type:   "push",
		Title:  req.Title,
		Body:   req.Body,
	}
	s.repo.Create(ctx, notif)

	// Get push tokens for user
	tokens, err := s.repo.GetPushTokens(ctx, req.UserID)
	if err != nil || len(tokens) == 0 {
		return nil // No tokens — silently skip
	}

	// Send via Expo Push API
	for _, pt := range tokens {
		go func(token string) {
			msg := map[string]interface{}{
				"to":    token,
				"title": req.Title,
				"body":  req.Body,
				"sound": "default",
			}
			if req.Data != nil {
				msg["data"] = req.Data
			}

			body, _ := json.Marshal(msg)
			resp, err := http.Post(expoPushURL, "application/json", bytes.NewReader(body))
			if err != nil {
				log.Printf("Expo push error: %v", err)
				return
			}
			resp.Body.Close()
			if resp.StatusCode != 200 {
				log.Printf("Expo push non-200: %d", resp.StatusCode)
			}
		}(pt.Token)
	}

	return nil
}

// CreateNotification stores a notification without pushing
func (s *NotificationService) CreateNotification(ctx context.Context, userID, nType, title, body string) error {
	notif := &models.Notification{
		ID:     uuid.New().String(),
		UserID: userID,
		Type:   nType,
		Title:  title,
		Body:   body,
	}
	return s.repo.Create(ctx, notif)
}

// SendToUser sends a push notification and stores it
func (s *NotificationService) SendToUser(ctx context.Context, userID, title, body string, data map[string]string) {
	s.SendPush(ctx, &models.SendPushRequest{
		UserID: userID,
		Title:  title,
		Body:   body,
		Data:   data,
	})
	fmt.Printf("NOTIF [%s]: %s - %s\n", userID, title, body)
}
