package service

import (
	"context"

	"github.com/teamsphere/server/internal/repository"
)

type AuditLogService struct {
	repo repository.AuditLogRepository
}

func NewAuditLogService(repo repository.AuditLogRepository) *AuditLogService {
	return &AuditLogService{repo: repo}
}

func (s *AuditLogService) Record(ctx context.Context, userID int64, action, entityType string, entityID int64, meta any, ip, userAgent string) error {
	return s.repo.Create(ctx, userID, action, entityType, entityID, meta, ip, userAgent)
}
