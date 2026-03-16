package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/teamsphere/server/internal/model"
)

type AuditLogRepo struct {
	db DBTX
}

func NewAuditLogRepo(db DBTX) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (r *AuditLogRepo) Create(ctx context.Context, userID int64, action, entityType string, entityID int64, meta any, ip, userAgent string) error {
	payload := "{}"
	if meta != nil {
		if raw, err := json.Marshal(meta); err == nil {
			payload = string(raw)
		}
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, meta, ip, user_agent)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
	`, userID, action, entityType, entityID, payload, ip, userAgent)
	if err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

func (r *AuditLogRepo) ListByUser(ctx context.Context, userID int64, limit int) ([]*model.AuditLog, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, action, entity_type, entity_id, meta::text, ip, user_agent, created_at
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*model.AuditLog
	for rows.Next() {
		l := &model.AuditLog{}
		if err := rows.Scan(&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID, &l.Meta, &l.IP, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit logs: %w", err)
	}
	return logs, nil
}

