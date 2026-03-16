package service

import (
	"context"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
)

type NotificationService struct {
	repo repository.NotificationRepository
}

func NewNotificationService(repo repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) Create(ctx context.Context, n *model.Notification) (*model.Notification, error) {
	return s.repo.Create(ctx, n)
}

func (s *NotificationService) List(ctx context.Context, userID int64, unreadOnly bool, limit int) ([]*model.Notification, error) {
	return s.repo.List(ctx, userID, unreadOnly, limit)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, id int64) error {
	return s.repo.MarkRead(ctx, userID, id)
}
