package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

type MessageReadRepo struct {
	db DBTX
}

func NewMessageReadRepo(db DBTX) *MessageReadRepo {
	return &MessageReadRepo{db: db}
}

func (r *MessageReadRepo) Upsert(ctx context.Context, userID, roomID, lastReadMsgID int64) (*model.MessageRead, error) {
	mr := &model.MessageRead{}
	err := r.db.QueryRow(ctx, `
		INSERT INTO message_reads (user_id, room_id, last_read_msg_id, read_at, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW(), NOW())
		ON CONFLICT (user_id, room_id)
		DO UPDATE SET last_read_msg_id = GREATEST(message_reads.last_read_msg_id, EXCLUDED.last_read_msg_id),
		              read_at = NOW(), updated_at = NOW()
		RETURNING user_id, room_id, last_read_msg_id, read_at, created_at, updated_at
	`, userID, roomID, lastReadMsgID).Scan(
		&mr.UserID, &mr.RoomID, &mr.LastReadMsgID, &mr.ReadAt, &mr.CreatedAt, &mr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert message read: %w", err)
	}
	return mr, nil
}

func (r *MessageReadRepo) Get(ctx context.Context, userID, roomID int64) (*model.MessageRead, error) {
	mr := &model.MessageRead{}
	err := r.db.QueryRow(ctx, `
		SELECT user_id, room_id, last_read_msg_id, read_at, created_at, updated_at
		FROM message_reads
		WHERE user_id = $1 AND room_id = $2
	`, userID, roomID).Scan(
		&mr.UserID, &mr.RoomID, &mr.LastReadMsgID, &mr.ReadAt, &mr.CreatedAt, &mr.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get message read: %w", err)
	}
	return mr, nil
}

func (r *MessageReadRepo) GetUnreadCount(ctx context.Context, userID, roomID int64, lastReadMsgID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM messages
		WHERE room_id = $1
		  AND id > $2
		  AND deleted_at IS NULL
	`, roomID, lastReadMsgID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread: %w", err)
	}
	return count, nil
}

func (r *MessageReadRepo) GetLastReadID(ctx context.Context, userID, roomID int64) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		SELECT last_read_msg_id FROM message_reads WHERE user_id = $1 AND room_id = $2
	`, userID, roomID).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("get last read id: %w", err)
	}
	return id, nil
}

func (r *MessageReadRepo) TouchIfMissing(ctx context.Context, userID, roomID, fallbackMsgID int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO message_reads (user_id, room_id, last_read_msg_id, read_at, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW(), NOW())
		ON CONFLICT (user_id, room_id) DO NOTHING
	`, userID, roomID, fallbackMsgID)
	if err != nil {
		return fmt.Errorf("touch message read: %w", err)
	}
	return nil
}

func (r *MessageReadRepo) MarkReadAtLatest(ctx context.Context, userID, roomID int64) (*model.MessageRead, error) {
	var lastID int64
	err := r.db.QueryRow(ctx, `
		SELECT COALESCE(MAX(id), 0) FROM messages WHERE room_id = $1
	`, roomID).Scan(&lastID)
	if err != nil {
		return nil, fmt.Errorf("get latest message id: %w", err)
	}
	return r.Upsert(ctx, userID, roomID, lastID)
}
