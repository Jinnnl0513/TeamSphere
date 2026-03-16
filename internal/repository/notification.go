package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

type NotificationRepo struct {
	db DBTX
}

func NewNotificationRepo(db DBTX) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(ctx context.Context, n *model.Notification) (*model.Notification, error) {
	out := &model.Notification{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO notifications (user_id, type, title, body, is_read, ref_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, user_id, type, title, body, is_read, ref_id, created_at
	`, n.UserID, n.Type, n.Title, n.Body, n.IsRead, n.RefID).Scan(
		&out.ID, &out.UserID, &out.Type, &out.Title, &out.Body, &out.IsRead, &out.RefID, &out.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return out, nil
}

func (r *NotificationRepo) List(ctx context.Context, userID int64, unreadOnly bool, limit int) ([]*model.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	query := `
		SELECT id, user_id, type, title, body, is_read, ref_id, created_at
		FROM notifications
		WHERE user_id = $1`
	if unreadOnly {
		query += " AND is_read = false"
	}
	query += " ORDER BY created_at DESC LIMIT $2"

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var items []*model.Notification
	for rows.Next() {
		n := &model.Notification{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.IsRead, &n.RefID, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		items = append(items, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return items, nil
}

func (r *NotificationRepo) MarkRead(ctx context.Context, userID, id int64) error {
	_, err := r.db.Exec(ctx, `UPDATE notifications SET is_read = true WHERE user_id = $1 AND id = $2`, userID, id)
	if err != nil {
		return fmt.Errorf("mark notification read: %w", err)
	}
	return nil
}

func (r *NotificationRepo) Get(ctx context.Context, userID, id int64) (*model.Notification, error) {
	n := &model.Notification{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, type, title, body, is_read, ref_id, created_at
		FROM notifications
		WHERE user_id = $1 AND id = $2
	`, userID, id).Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &n.IsRead, &n.RefID, &n.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get notification: %w", err)
	}
	return n, nil
}

