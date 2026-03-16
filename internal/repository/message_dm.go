package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

// DMWithUser is a DM with sender info and optional reply_to snapshot attached.
type DMWithUser struct {
	model.DirectMessage
	User    *model.UserInfo  `json:"user"`
	ReplyTo *model.ReplyInfo `json:"reply_to,omitempty"`
}

// Conversation represents a recent DM conversation.
type Conversation struct {
	User      model.UserInfo `json:"user"`
	LastMsg   string         `json:"last_message"`
	LastMsgAt *time.Time     `json:"last_message_at"`
}

// CountDMs returns the total number of direct messages.
func (r *MessageRepo) CountDMs(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM direct_messages`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count direct messages: %w", err)
	}
	return count, nil
}

// CreateDM inserts a direct message. clientMsgID and replyToID are optional.
func (r *MessageRepo) CreateDM(ctx context.Context, content string, senderID, receiverID int64, msgType string, fileSize *int64, mimeType *string, clientMsgID *string, replyToID *int64, forwardMeta *model.ForwardInfo) (*model.DirectMessage, error) {
	dm := &model.DirectMessage{}
	forwardJSON, err := marshalForwardMeta(forwardMeta)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRow(ctx,
		`INSERT INTO direct_messages (content, sender_id, receiver_id, msg_type, file_size, mime_type, client_msg_id, reply_to_id, forward_meta)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, content, sender_id, receiver_id, msg_type, file_size, mime_type, reply_to_id, forward_meta, deleted_at, client_msg_id, read_at, created_at, updated_at`,
		content, senderID, receiverID, msgType, fileSize, mimeType, clientMsgID, replyToID, forwardJSON,
	).Scan(&dm.ID, &dm.Content, &dm.SenderID, &dm.ReceiverID, &dm.MsgType,
		&dm.FileSize, &dm.MimeType, &dm.ReplyToID, &forwardJSON, &dm.DeletedAt, &dm.ClientMsgID, &dm.ReadAt, &dm.CreatedAt, &dm.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create dm: %w", err)
	}
	dm.ForwardMeta = unmarshalForwardMeta(forwardJSON)
	return dm, nil
}

// GetDMByClientMsgID retrieves a DM by client_msg_id (for dedup).
func (r *MessageRepo) GetDMByClientMsgID(ctx context.Context, clientMsgID string) (*model.DirectMessage, error) {
	dm := &model.DirectMessage{}
	var forwardJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, content, sender_id, receiver_id, msg_type, file_size, mime_type, reply_to_id, forward_meta, deleted_at, client_msg_id, read_at, created_at, updated_at
		 FROM direct_messages WHERE client_msg_id = $1`, clientMsgID,
	).Scan(&dm.ID, &dm.Content, &dm.SenderID, &dm.ReceiverID, &dm.MsgType,
		&dm.FileSize, &dm.MimeType, &dm.ReplyToID, &forwardJSON, &dm.DeletedAt, &dm.ClientMsgID, &dm.ReadAt, &dm.CreatedAt, &dm.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get dm by client_msg_id: %w", err)
	}
	dm.ForwardMeta = unmarshalForwardMeta(forwardJSON)
	return dm, nil
}

// GetDMByID retrieves a direct message by ID.
func (r *MessageRepo) GetDMByID(ctx context.Context, id int64) (*model.DirectMessage, error) {
	dm := &model.DirectMessage{}
	var forwardJSON []byte
	err := r.db.QueryRow(ctx,
		`SELECT id, content, sender_id, receiver_id, msg_type, file_size, mime_type, reply_to_id, forward_meta, deleted_at, client_msg_id, read_at, created_at, updated_at
		 FROM direct_messages WHERE id = $1`, id,
	).Scan(&dm.ID, &dm.Content, &dm.SenderID, &dm.ReceiverID, &dm.MsgType,
		&dm.FileSize, &dm.MimeType, &dm.ReplyToID, &forwardJSON, &dm.DeletedAt, &dm.ClientMsgID, &dm.ReadAt, &dm.CreatedAt, &dm.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get dm: %w", err)
	}
	dm.ForwardMeta = unmarshalForwardMeta(forwardJSON)
	return dm, nil
}

// SoftDeleteDM marks a direct message as deleted (soft delete for recall).
func (r *MessageRepo) SoftDeleteDM(ctx context.Context, msgID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE direct_messages SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`, msgID)
	if err != nil {
		return fmt.Errorf("soft delete dm: %w", err)
	}
	return nil
}

// UpdateDMContent updates a direct message's content and returns the updated_at timestamp.
func (r *MessageRepo) UpdateDMContent(ctx context.Context, msgID int64, content string) (time.Time, error) {
	var updatedAt time.Time
	err := r.db.QueryRow(ctx,
		`UPDATE direct_messages SET content = $1, updated_at = NOW() WHERE id = $2 RETURNING updated_at`,
		content, msgID,
	).Scan(&updatedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("update dm content: %w", err)
	}
	return updatedAt, nil
}

// ListDMs returns direct messages between two users with cursor pagination.
func (r *MessageRepo) ListDMs(ctx context.Context, userA, userB int64, beforeID, afterID int64, limit int) ([]*DMWithUser, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	var query string
	var args []any

	baseSelect := `
		SELECT d.id, d.content, d.sender_id, d.receiver_id, d.msg_type, d.file_size, d.mime_type, d.reply_to_id, d.forward_meta, d.deleted_at, d.client_msg_id, d.read_at, d.created_at, d.updated_at,
		       u.id,  u.username,  u.avatar_url,
		       rd.id, rd.content, rd.msg_type, rd.deleted_at,
		       ru.id, ru.username, ru.avatar_url
		FROM direct_messages d
		LEFT JOIN users u  ON u.id  = d.sender_id
		LEFT JOIN direct_messages rd ON rd.id = d.reply_to_id
		LEFT JOIN users ru ON ru.id = rd.sender_id`

	pairFilter := ` WHERE ((d.sender_id = $1 AND d.receiver_id = $2) OR (d.sender_id = $2 AND d.receiver_id = $1))`

	if afterID > 0 {
		query = baseSelect + pairFilter + ` AND d.id > $3 ORDER BY d.id ASC  LIMIT $4`
		args = []any{userA, userB, afterID, limit}
	} else if beforeID > 0 {
		query = baseSelect + pairFilter + ` AND d.id < $3 ORDER BY d.id DESC LIMIT $4`
		args = []any{userA, userB, beforeID, limit}
	} else {
		query = baseSelect + pairFilter + ` ORDER BY d.id DESC LIMIT $3`
		args = []any{userA, userB, limit}
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list dms: %w", err)
	}
	defer rows.Close()

	var messages []*DMWithUser
	for rows.Next() {
		dw := &DMWithUser{}
		var userID *int64
		var username, avatarURL *string
		var rtID *int64
		var rtContent *string
		var rtMsgType *string
		var rtDeletedAt *time.Time
		var ruID *int64
		var ruUsername, ruAvatarURL *string
		var forwardJSON []byte

		if err := rows.Scan(
			&dw.ID, &dw.Content, &dw.SenderID, &dw.ReceiverID, &dw.MsgType,
			&dw.FileSize, &dw.MimeType, &dw.ReplyToID, &forwardJSON, &dw.DeletedAt, &dw.ClientMsgID, &dw.ReadAt, &dw.CreatedAt, &dw.UpdatedAt,
			&userID, &username, &avatarURL,
			&rtID, &rtContent, &rtMsgType, &rtDeletedAt,
			&ruID, &ruUsername, &ruAvatarURL,
		); err != nil {
			return nil, fmt.Errorf("scan dm: %w", err)
		}

		dw.ForwardMeta = unmarshalForwardMeta(forwardJSON)

		if userID != nil {
			dw.User = &model.UserInfo{
				ID:        *userID,
				Username:  derefStr(username),
				AvatarURL: derefStr(avatarURL),
			}
		}

		if rtID != nil {
			rt := &model.ReplyInfo{
				ID:        *rtID,
				MsgType:   derefStr(rtMsgType),
				IsDeleted: rtDeletedAt != nil,
			}
			if rtDeletedAt == nil {
				rt.Content = truncateRunes(derefStr(rtContent), 200)
			}
			if ruID != nil {
				rt.User = model.UserInfo{
					ID:        *ruID,
					Username:  derefStr(ruUsername),
					AvatarURL: derefStr(ruAvatarURL),
				}
			}
			dw.ReplyTo = rt
		}

		messages = append(messages, dw)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dms: %w", err)
	}

	// For DESC queries, reverse to ascending order
	if afterID == 0 && len(messages) > 0 {
		for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
			messages[i], messages[j] = messages[j], messages[i]
		}
	}

	return messages, nil
}

// ListConversations returns the most recent DM conversations for a user.
func (r *MessageRepo) ListConversations(ctx context.Context, userID int64) ([]*Conversation, error) {
	rows, err := r.db.Query(ctx,
		`SELECT sub.peer_id, sub.content, sub.created_at,
		        u.id, u.username, u.avatar_url
		 FROM (
			SELECT DISTINCT ON (peer_id) peer_id, d.content, d.created_at
			FROM (
				SELECT id, content, created_at,
				       CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END AS peer_id
				FROM direct_messages
				WHERE (sender_id = $1 OR receiver_id = $1) AND deleted_at IS NULL
			) d
			ORDER BY peer_id, d.created_at DESC
		 ) sub
		 JOIN users u ON u.id = sub.peer_id
		 ORDER BY sub.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var convos []*Conversation
	for rows.Next() {
		c := &Conversation{}
		var peerID int64
		if err := rows.Scan(
			&peerID, &c.LastMsg, &c.LastMsgAt,
			&c.User.ID, &c.User.Username, &c.User.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		convos = append(convos, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate conversations: %w", err)
	}

	return convos, nil
}

// MarkDMRead marks direct messages from peer as read up to lastReadMsgID (if provided).
func (r *MessageRepo) MarkDMRead(ctx context.Context, receiverID, peerID int64, lastReadMsgID int64) (int64, *time.Time, error) {
	var rows pgx.Rows
	var err error
	if lastReadMsgID > 0 {
		rows, err = r.db.Query(ctx,
			`UPDATE direct_messages
			 SET read_at = NOW()
			 WHERE receiver_id = $1 AND sender_id = $2 AND read_at IS NULL AND id <= $3
			 RETURNING read_at`, receiverID, peerID, lastReadMsgID,
		)
	} else {
		rows, err = r.db.Query(ctx,
			`UPDATE direct_messages
			 SET read_at = NOW()
			 WHERE receiver_id = $1 AND sender_id = $2 AND read_at IS NULL
			 RETURNING read_at`, receiverID, peerID,
		)
	}
	if err != nil {
		return 0, nil, fmt.Errorf("mark dm read: %w", err)
	}
	defer rows.Close()

	var count int64
	var lastReadAt *time.Time
	for rows.Next() {
		var ts time.Time
		if err := rows.Scan(&ts); err != nil {
			return 0, nil, fmt.Errorf("scan dm read: %w", err)
		}
		count++
		lastReadAt = &ts
	}
	if err := rows.Err(); err != nil {
		return 0, nil, fmt.Errorf("iterate dm read: %w", err)
	}
	return count, lastReadAt, nil
}
