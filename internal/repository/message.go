package repository

import (
    "context"
    "errors"
    "fmt"
    "time"

    "github.com/teamsphere/server/internal/model"
    "github.com/jackc/pgx/v5"
)

type MessageRepo struct {
    db DBTX
}

func NewMessageRepo(db DBTX) *MessageRepo {
    return &MessageRepo{db: db}
}

// CountMessages returns the total number of room messages.
func (r *MessageRepo) CountMessages(ctx context.Context) (int64, error) {
    var count int64
    if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM messages`).Scan(&count); err != nil {
        return 0, fmt.Errorf("count messages: %w", err)
    }
    return count, nil
}

// Create inserts a room message. clientMsgID and replyToID are optional.
func (r *MessageRepo) Create(ctx context.Context, content string, userID, roomID int64, msgType string, mentions []int64, fileSize *int64, mimeType *string, clientMsgID *string, replyToID *int64, forwardMeta *model.ForwardInfo) (*model.Message, error) {
    if mentions == nil {
        mentions = []int64{}
    }

    msg := &model.Message{}
    forwardJSON, err := marshalForwardMeta(forwardMeta)
    if err != nil {
        return nil, err
    }
    err = r.db.QueryRow(ctx,
        `INSERT INTO messages (content, user_id, room_id, msg_type, mentions, file_size, mime_type, client_msg_id, reply_to_id, forward_meta)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
         RETURNING id, content, user_id, room_id, msg_type, mentions, file_size, mime_type, reply_to_id, forward_meta, deleted_at, client_msg_id, created_at, updated_at`,
        content, userID, roomID, msgType, mentions, fileSize, mimeType, clientMsgID, replyToID, forwardJSON,
    ).Scan(&msg.ID, &msg.Content, &msg.UserID, &msg.RoomID, &msg.MsgType, &msg.Mentions,
        &msg.FileSize, &msg.MimeType, &msg.ReplyToID, &forwardJSON, &msg.DeletedAt, &msg.ClientMsgID, &msg.CreatedAt, &msg.UpdatedAt)
    if err != nil {
        return nil, fmt.Errorf("create message: %w", err)
    }
    msg.ForwardMeta = unmarshalForwardMeta(forwardJSON)
    return msg, nil
}

// GetByID retrieves a message by ID.
func (r *MessageRepo) GetByID(ctx context.Context, id int64) (*model.Message, error) {
    msg := &model.Message{}
    var forwardJSON []byte
    err := r.db.QueryRow(ctx,
        `SELECT id, content, user_id, room_id, msg_type, mentions, file_size, mime_type, reply_to_id, forward_meta, deleted_at, client_msg_id, created_at, updated_at
         FROM messages WHERE id = $1`, id,
    ).Scan(&msg.ID, &msg.Content, &msg.UserID, &msg.RoomID, &msg.MsgType, &msg.Mentions,
        &msg.FileSize, &msg.MimeType, &msg.ReplyToID, &forwardJSON, &msg.DeletedAt, &msg.ClientMsgID, &msg.CreatedAt, &msg.UpdatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, nil
        }
        return nil, fmt.Errorf("get message: %w", err)
    }
    msg.ForwardMeta = unmarshalForwardMeta(forwardJSON)
    return msg, nil
}

// GetByClientMsgID retrieves a message by client_msg_id (for dedup).
func (r *MessageRepo) GetByClientMsgID(ctx context.Context, clientMsgID string) (*model.Message, error) {
    msg := &model.Message{}
    var forwardJSON []byte
    err := r.db.QueryRow(ctx,
        `SELECT id, content, user_id, room_id, msg_type, mentions, file_size, mime_type, reply_to_id, forward_meta, deleted_at, client_msg_id, created_at, updated_at
         FROM messages WHERE client_msg_id = $1`, clientMsgID,
    ).Scan(&msg.ID, &msg.Content, &msg.UserID, &msg.RoomID, &msg.MsgType, &msg.Mentions,
        &msg.FileSize, &msg.MimeType, &msg.ReplyToID, &forwardJSON, &msg.DeletedAt, &msg.ClientMsgID, &msg.CreatedAt, &msg.UpdatedAt)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, nil
        }
        return nil, fmt.Errorf("get message by client_msg_id: %w", err)
    }
    msg.ForwardMeta = unmarshalForwardMeta(forwardJSON)
    return msg, nil
}

// MessageWithUser is a message with sender info and optional reply_to snapshot attached.
type MessageWithUser struct {
    model.Message
    User    *model.UserInfo  `json:"user"`
    ReplyTo *model.ReplyInfo `json:"reply_to,omitempty"`
}

// ListByRoom returns messages for a room with cursor pagination.
// beforeID: return messages with ID < beforeID (older messages).
// afterID: return messages with ID > afterID (newer messages).
// limit: max messages to return (default 50, max 100).
// Only one of beforeID/afterID should be non-zero.
func (r *MessageRepo) ListByRoom(ctx context.Context, roomID int64, beforeID, afterID int64, limit int) ([]*MessageWithUser, error) {
    if limit <= 0 {
        limit = 50
    }
    if limit > 100 {
        limit = 100
    }

    var query string
    var args []any

    // Base SELECT: main message columns + sender + reply_to snapshot (LEFT JOIN twice)
    baseSelect := `
        SELECT m.id, m.content, m.user_id, m.room_id, m.msg_type, m.mentions, m.file_size, m.mime_type, m.reply_to_id, m.forward_meta, m.deleted_at, m.client_msg_id, m.created_at, m.updated_at,
               u.id, u.username, u.avatar_url,
               rm.id, rm.content, rm.msg_type, rm.deleted_at,
               ru.id, ru.username, ru.avatar_url
        FROM messages m
        LEFT JOIN users u  ON u.id  = m.user_id
        LEFT JOIN messages rm ON rm.id = m.reply_to_id
        LEFT JOIN users ru ON ru.id = rm.user_id`

    if afterID > 0 {
        query = baseSelect + ` WHERE m.room_id = $1 AND m.id > $2 ORDER BY m.id ASC  LIMIT $3`
        args = []any{roomID, afterID, limit}
    } else if beforeID > 0 {
        query = baseSelect + ` WHERE m.room_id = $1 AND m.id < $2 ORDER BY m.id DESC LIMIT $3`
        args = []any{roomID, beforeID, limit}
    } else {
        query = baseSelect + ` WHERE m.room_id = $1                  ORDER BY m.id DESC LIMIT $2`
        args = []any{roomID, limit}
    }

    rows, err := r.db.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("list messages: %w", err)
    }
    defer rows.Close()

    var messages []*MessageWithUser
    for rows.Next() {
        mw := &MessageWithUser{}
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
            &mw.ID, &mw.Content, &mw.UserID, &mw.RoomID, &mw.MsgType, &mw.Mentions,
            &mw.FileSize, &mw.MimeType, &mw.ReplyToID, &forwardJSON, &mw.DeletedAt, &mw.ClientMsgID, &mw.CreatedAt, &mw.UpdatedAt,
            &userID, &username, &avatarURL,
            &rtID, &rtContent, &rtMsgType, &rtDeletedAt,
            &ruID, &ruUsername, &ruAvatarURL,
        ); err != nil {
            return nil, fmt.Errorf("scan message: %w", err)
        }

        mw.ForwardMeta = unmarshalForwardMeta(forwardJSON)

        if userID != nil {
            mw.User = &model.UserInfo{
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
            mw.ReplyTo = rt
        }

        messages = append(messages, mw)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate messages: %w", err)
    }

    // For DESC queries (beforeID or latest), reverse to ascending order
    if afterID == 0 && len(messages) > 0 {
        for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
            messages[i], messages[j] = messages[j], messages[i]
        }
    }

    return messages, nil
}

// ListThreadByRoom returns thread messages for a room and root message with cursor pagination.
// beforeID: return messages with ID < beforeID (older messages).
// afterID: return messages with ID > afterID (newer messages).
// limit: max messages to return (default 50, max 100).
// Only one of beforeID/afterID should be non-zero.
func (r *MessageRepo) ListThreadByRoom(ctx context.Context, roomID, rootMsgID int64, beforeID, afterID int64, limit int) ([]*MessageWithUser, error) {
    if limit <= 0 {
        limit = 50
    }
    if limit > 100 {
        limit = 100
    }

    var query string
    var args []any

    baseSelect := `
        SELECT m.id, m.content, m.user_id, m.room_id, m.msg_type, m.mentions, m.file_size, m.mime_type, m.reply_to_id, m.forward_meta, m.deleted_at, m.client_msg_id, m.created_at, m.updated_at,
               u.id, u.username, u.avatar_url,
               rm.id, rm.content, rm.msg_type, rm.deleted_at,
               ru.id, ru.username, ru.avatar_url
        FROM messages m
        LEFT JOIN users u  ON u.id  = m.user_id
        LEFT JOIN messages rm ON rm.id = m.reply_to_id
        LEFT JOIN users ru ON ru.id = rm.user_id`

    if afterID > 0 {
        query = baseSelect + ` WHERE m.room_id = $1 AND m.reply_to_id = $2 AND m.id > $3 ORDER BY m.id ASC  LIMIT $4`
        args = []any{roomID, rootMsgID, afterID, limit}
    } else if beforeID > 0 {
        query = baseSelect + ` WHERE m.room_id = $1 AND m.reply_to_id = $2 AND m.id < $3 ORDER BY m.id DESC LIMIT $4`
        args = []any{roomID, rootMsgID, beforeID, limit}
    } else {
        query = baseSelect + ` WHERE m.room_id = $1 AND m.reply_to_id = $2 ORDER BY m.id ASC LIMIT $3`
        args = []any{roomID, rootMsgID, limit}
    }

    rows, err := r.db.Query(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("list thread messages: %w", err)
    }
    defer rows.Close()

    var messages []*MessageWithUser
    for rows.Next() {
        mw := &MessageWithUser{}
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
            &mw.ID, &mw.Content, &mw.UserID, &mw.RoomID, &mw.MsgType, &mw.Mentions,
            &mw.FileSize, &mw.MimeType, &mw.ReplyToID, &forwardJSON, &mw.DeletedAt, &mw.ClientMsgID, &mw.CreatedAt, &mw.UpdatedAt,
            &userID, &username, &avatarURL,
            &rtID, &rtContent, &rtMsgType, &rtDeletedAt,
            &ruID, &ruUsername, &ruAvatarURL,
        ); err != nil {
            return nil, fmt.Errorf("scan thread message: %w", err)
        }

        mw.ForwardMeta = unmarshalForwardMeta(forwardJSON)

        if userID != nil {
            mw.User = &model.UserInfo{
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
            mw.ReplyTo = rt
        }

        messages = append(messages, mw)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate thread messages: %w", err)
    }

    if afterID == 0 && beforeID > 0 && len(messages) > 0 {
        for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
            messages[i], messages[j] = messages[j], messages[i]
        }
    }

    return messages, nil
}

// PinMessage pins a message in a room (upsert).
func (r *MessageRepo) PinMessage(ctx context.Context, roomID, msgID, userID int64) error {
    _, err := r.db.Exec(ctx, `
        INSERT INTO pinned_messages (room_id, msg_id, pinned_by)
        VALUES ($1, $2, $3)
        ON CONFLICT (room_id, msg_id) DO UPDATE SET pinned_by = EXCLUDED.pinned_by, created_at = NOW()`,
        roomID, msgID, userID)
    if err != nil {
        return fmt.Errorf("pin message: %w", err)
    }
    return nil
}

// UnpinMessage removes a pinned message from a room.
func (r *MessageRepo) UnpinMessage(ctx context.Context, roomID, msgID int64) error {
    _, err := r.db.Exec(ctx, `DELETE FROM pinned_messages WHERE room_id = $1 AND msg_id = $2`, roomID, msgID)
    if err != nil {
        return fmt.Errorf("unpin message: %w", err)
    }
    return nil
}

// ListPinnedMessages returns pinned messages for a room (most recent first).
func (r *MessageRepo) ListPinnedMessages(ctx context.Context, roomID int64) ([]*MessageWithUser, error) {
    query := `
        SELECT m.id, m.content, m.user_id, m.room_id, m.msg_type, m.mentions, m.file_size, m.mime_type, m.reply_to_id, m.forward_meta, m.deleted_at, m.client_msg_id, m.created_at, m.updated_at,
               u.id, u.username, u.avatar_url,
               rm.id, rm.content, rm.msg_type, rm.deleted_at,
               ru.id, ru.username, ru.avatar_url
        FROM pinned_messages p
        JOIN messages m ON m.id = p.msg_id
        LEFT JOIN users u  ON u.id  = m.user_id
        LEFT JOIN messages rm ON rm.id = m.reply_to_id
        LEFT JOIN users ru ON ru.id = rm.user_id
        WHERE p.room_id = $1
        ORDER BY p.created_at DESC`

    rows, err := r.db.Query(ctx, query, roomID)
    if err != nil {
        return nil, fmt.Errorf("list pinned messages: %w", err)
    }
    defer rows.Close()

    var messages []*MessageWithUser
    for rows.Next() {
        mw := &MessageWithUser{}
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
            &mw.ID, &mw.Content, &mw.UserID, &mw.RoomID, &mw.MsgType, &mw.Mentions,
            &mw.FileSize, &mw.MimeType, &mw.ReplyToID, &forwardJSON, &mw.DeletedAt, &mw.ClientMsgID, &mw.CreatedAt, &mw.UpdatedAt,
            &userID, &username, &avatarURL,
            &rtID, &rtContent, &rtMsgType, &rtDeletedAt,
            &ruID, &ruUsername, &ruAvatarURL,
        ); err != nil {
            return nil, fmt.Errorf("scan pinned message: %w", err)
        }

        mw.ForwardMeta = unmarshalForwardMeta(forwardJSON)

        if userID != nil {
            mw.User = &model.UserInfo{
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
            mw.ReplyTo = rt
        }

        messages = append(messages, mw)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate pinned messages: %w", err)
    }

    return messages, nil
}

// SearchMessages searches room messages with optional filters.
func (r *MessageRepo) SearchMessages(ctx context.Context, userID int64, query string, roomID, senderID int64, from, to *time.Time, limit int) ([]*MessageWithUser, error) {
    if limit <= 0 {
        limit = 50
    }
    if limit > 100 {
        limit = 100
    }

    baseSelect := `
        SELECT m.id, m.content, m.user_id, m.room_id, m.msg_type, m.mentions, m.file_size, m.mime_type, m.reply_to_id, m.forward_meta, m.deleted_at, m.client_msg_id, m.created_at, m.updated_at,
               u.id, u.username, u.avatar_url,
               rm.id, rm.content, rm.msg_type, rm.deleted_at,
               ru.id, ru.username, ru.avatar_url
        FROM messages m
        LEFT JOIN users u  ON u.id  = m.user_id
        LEFT JOIN messages rm ON rm.id = m.reply_to_id
        LEFT JOIN users ru ON ru.id = rm.user_id`

    clauses := []string{"m.deleted_at IS NULL"}
    args := []any{}

    if roomID > 0 {
        args = append(args, roomID)
        clauses = append(clauses, fmt.Sprintf("m.room_id = $%d", len(args)))
    }
    if senderID > 0 {
        args = append(args, senderID)
        clauses = append(clauses, fmt.Sprintf("m.user_id = $%d", len(args)))
    }
    if from != nil {
        args = append(args, *from)
        clauses = append(clauses, fmt.Sprintf("m.created_at >= $%d", len(args)))
    }
    if to != nil {
        args = append(args, *to)
        clauses = append(clauses, fmt.Sprintf("m.created_at <= $%d", len(args)))
    }
    if query != "" {
        args = append(args, query)
        clauses = append(clauses, fmt.Sprintf("to_tsvector('simple', m.content) @@ plainto_tsquery('simple', $%d)", len(args)))
    }

    args = append(args, limit)
    clausesSQL := " WHERE " + joinClauses(clauses, " AND ")
    querySQL := baseSelect + clausesSQL + fmt.Sprintf(" ORDER BY m.created_at DESC LIMIT $%d", len(args))

    rows, err := r.db.Query(ctx, querySQL, args...)
    if err != nil {
        return nil, fmt.Errorf("search messages: %w", err)
    }
    defer rows.Close()

    var messages []*MessageWithUser
    for rows.Next() {
        mw := &MessageWithUser{}
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
            &mw.ID, &mw.Content, &mw.UserID, &mw.RoomID, &mw.MsgType, &mw.Mentions,
            &mw.FileSize, &mw.MimeType, &mw.ReplyToID, &forwardJSON, &mw.DeletedAt, &mw.ClientMsgID, &mw.CreatedAt, &mw.UpdatedAt,
            &userID, &username, &avatarURL,
            &rtID, &rtContent, &rtMsgType, &rtDeletedAt,
            &ruID, &ruUsername, &ruAvatarURL,
        ); err != nil {
            return nil, fmt.Errorf("scan search message: %w", err)
        }

        mw.ForwardMeta = unmarshalForwardMeta(forwardJSON)

        if userID != nil {
            mw.User = &model.UserInfo{
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
            mw.ReplyTo = rt
        }

        messages = append(messages, mw)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate search messages: %w", err)
    }

    return messages, nil
}

// SoftDeleteBatch marks multiple room messages as deleted and returns affected IDs.
func (r *MessageRepo) SoftDeleteBatch(ctx context.Context, roomID int64, msgIDs []int64) ([]int64, error) {
    if len(msgIDs) == 0 {
        return []int64{}, nil
    }
    rows, err := r.db.Query(ctx,
        `UPDATE messages
         SET deleted_at = NOW(), updated_at = NOW()
         WHERE room_id = $1 AND id = ANY($2) AND deleted_at IS NULL
         RETURNING id`,
        roomID, msgIDs,
    )
    if err != nil {
        return nil, fmt.Errorf("soft delete batch: %w", err)
    }
    defer rows.Close()

    var ids []int64
    for rows.Next() {
        var id int64
        if err := rows.Scan(&id); err != nil {
            return nil, fmt.Errorf("scan deleted id: %w", err)
        }
        ids = append(ids, id)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate deleted ids: %w", err)
    }
    return ids, nil
}

// SoftDelete marks a message as deleted (soft delete for recall).
func (r *MessageRepo) SoftDelete(ctx context.Context, msgID int64) error {
    _, err := r.db.Exec(ctx,
        `UPDATE messages SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`, msgID)
    if err != nil {
        return fmt.Errorf("soft delete message: %w", err)
    }
    return nil
}

// UpdateContent updates a room message's content and returns the updated_at timestamp.
func (r *MessageRepo) UpdateContent(ctx context.Context, msgID int64, content string) (time.Time, error) {
    var updatedAt time.Time
    err := r.db.QueryRow(ctx,
        `UPDATE messages SET content = $1, updated_at = NOW() WHERE id = $2 RETURNING updated_at`,
        content, msgID,
    ).Scan(&updatedAt)
    if err != nil {
        return time.Time{}, fmt.Errorf("update message content: %w", err)
    }
    return updatedAt, nil
}

// GetRoomIDByMsgID returns the room_id for a given message ID.
func (r *MessageRepo) GetRoomIDByMsgID(ctx context.Context, msgID int64) (int64, error) {
    var roomID int64
    err := r.db.QueryRow(ctx, `SELECT room_id FROM messages WHERE id = $1`, msgID).Scan(&roomID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return 0, nil
        }
        return 0, fmt.Errorf("get room_id by msg_id: %w", err)
    }
    return roomID, nil
}
