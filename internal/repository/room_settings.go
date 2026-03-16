package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

type RoomSettingsRepo struct {
	db DBTX
}

func NewRoomSettingsRepo(db DBTX) *RoomSettingsRepo {
	return &RoomSettingsRepo{db: db}
}

func (r *RoomSettingsRepo) GetByRoomID(ctx context.Context, roomID int64) (*model.RoomSettings, error) {
	var s model.RoomSettings
	err := r.db.QueryRow(ctx, `
		SELECT id, room_id, is_public, require_approval, read_only, archived, topic, avatar_url,
		       slow_mode_seconds, message_retention_days, content_filter_mode, blocked_keywords,
		       allowed_link_domains, blocked_link_domains, allowed_file_types, max_file_size_mb,
		       pin_limit, notify_mode, notify_keywords, dnd_start, dnd_end,
		       anti_spam_rate, anti_spam_window_sec, anti_repeat, stats_enabled,
		       created_at, updated_at
		FROM room_settings WHERE room_id = $1`, roomID,
	).Scan(
		&s.ID, &s.RoomID, &s.IsPublic, &s.RequireApproval, &s.ReadOnly, &s.Archived, &s.Topic, &s.AvatarURL,
		&s.SlowModeSeconds, &s.MessageRetention, &s.ContentFilterMode, &s.BlockedKeywords,
		&s.AllowedLinkDomains, &s.BlockedLinkDomains, &s.AllowedFileTypes, &s.MaxFileSizeMB,
		&s.PinLimit, &s.NotifyMode, &s.NotifyKeywords, &s.DNDStart, &s.DNDEnd,
		&s.AntiSpamRate, &s.AntiSpamWindowSec, &s.AntiRepeat, &s.StatsEnabled,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get room settings: %w", err)
	}
	return &s, nil
}

func (r *RoomSettingsRepo) CreateDefault(ctx context.Context, roomID int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO room_settings (room_id)
		VALUES ($1)
		ON CONFLICT (room_id) DO NOTHING`, roomID)
	if err != nil {
		return fmt.Errorf("create default room settings: %w", err)
	}

	// Default role permissions
	defaults := []struct {
		role             string
		canSend          bool
		canUpload        bool
		canPin           bool
		canManageMembers bool
		canManageSettings bool
		canManageMessages bool
		canMentionAll    bool
	}{
		{"owner", true, true, true, true, true, true, true},
		{"admin", true, true, true, true, true, true, true},
		{"member", true, true, false, false, false, false, false},
	}
	for _, d := range defaults {
		_, _ = r.db.Exec(ctx, `
			INSERT INTO room_role_permissions (room_id, role, can_send, can_upload, can_pin, can_manage_members, can_manage_settings, can_manage_messages, can_mention_all)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (room_id, role) DO NOTHING`,
			roomID, d.role, d.canSend, d.canUpload, d.canPin, d.canManageMembers, d.canManageSettings, d.canManageMessages, d.canMentionAll,
		)
	}
	return nil
}

func (r *RoomSettingsRepo) Update(ctx context.Context, roomID int64, s *model.RoomSettings) (*model.RoomSettings, error) {
	var out model.RoomSettings
	err := r.db.QueryRow(ctx, `
		UPDATE room_settings SET
			is_public = $1,
			require_approval = $2,
			read_only = $3,
			archived = $4,
			topic = $5,
			avatar_url = $6,
			slow_mode_seconds = $7,
			message_retention_days = $8,
			content_filter_mode = $9,
			blocked_keywords = $10,
			allowed_link_domains = $11,
			blocked_link_domains = $12,
			allowed_file_types = $13,
			max_file_size_mb = $14,
			pin_limit = $15,
			notify_mode = $16,
			notify_keywords = $17,
			dnd_start = $18,
			dnd_end = $19,
			anti_spam_rate = $20,
			anti_spam_window_sec = $21,
			anti_repeat = $22,
			stats_enabled = $23,
			updated_at = NOW()
		WHERE room_id = $24
		RETURNING id, room_id, is_public, require_approval, read_only, archived, topic, avatar_url,
		          slow_mode_seconds, message_retention_days, content_filter_mode, blocked_keywords,
		          allowed_link_domains, blocked_link_domains, allowed_file_types, max_file_size_mb,
		          pin_limit, notify_mode, notify_keywords, dnd_start, dnd_end,
		          anti_spam_rate, anti_spam_window_sec, anti_repeat, stats_enabled,
		          created_at, updated_at`,
		s.IsPublic, s.RequireApproval, s.ReadOnly, s.Archived, s.Topic, s.AvatarURL,
		s.SlowModeSeconds, s.MessageRetention, s.ContentFilterMode, s.BlockedKeywords,
		s.AllowedLinkDomains, s.BlockedLinkDomains, s.AllowedFileTypes, s.MaxFileSizeMB,
		s.PinLimit, s.NotifyMode, s.NotifyKeywords, s.DNDStart, s.DNDEnd,
		s.AntiSpamRate, s.AntiSpamWindowSec, s.AntiRepeat, s.StatsEnabled,
		roomID,
	).Scan(
		&out.ID, &out.RoomID, &out.IsPublic, &out.RequireApproval, &out.ReadOnly, &out.Archived, &out.Topic, &out.AvatarURL,
		&out.SlowModeSeconds, &out.MessageRetention, &out.ContentFilterMode, &out.BlockedKeywords,
		&out.AllowedLinkDomains, &out.BlockedLinkDomains, &out.AllowedFileTypes, &out.MaxFileSizeMB,
		&out.PinLimit, &out.NotifyMode, &out.NotifyKeywords, &out.DNDStart, &out.DNDEnd,
		&out.AntiSpamRate, &out.AntiSpamWindowSec, &out.AntiRepeat, &out.StatsEnabled,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update room settings: %w", err)
	}
	return &out, nil
}

func (r *RoomSettingsRepo) ListPermissions(ctx context.Context, roomID int64) ([]*model.RoomRolePermission, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, room_id, role, can_send, can_upload, can_pin, can_manage_members, can_manage_settings, can_manage_messages, can_mention_all, created_at, updated_at
		FROM room_role_permissions WHERE room_id = $1 ORDER BY role`, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room permissions: %w", err)
	}
	defer rows.Close()

	var items []*model.RoomRolePermission
	for rows.Next() {
		var p model.RoomRolePermission
		if err := rows.Scan(&p.ID, &p.RoomID, &p.Role, &p.CanSend, &p.CanUpload, &p.CanPin, &p.CanManageMembers, &p.CanManageSettings, &p.CanManageMessages, &p.CanMentionAll, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan room permission: %w", err)
		}
		items = append(items, &p)
	}
	return items, nil
}

func (r *RoomSettingsRepo) GetPermission(ctx context.Context, roomID int64, role string) (*model.RoomRolePermission, error) {
	var p model.RoomRolePermission
	err := r.db.QueryRow(ctx, `
		SELECT id, room_id, role, can_send, can_upload, can_pin, can_manage_members, can_manage_settings, can_manage_messages, can_mention_all, created_at, updated_at
		FROM room_role_permissions WHERE room_id = $1 AND role = $2`, roomID, role,
	).Scan(&p.ID, &p.RoomID, &p.Role, &p.CanSend, &p.CanUpload, &p.CanPin, &p.CanManageMembers, &p.CanManageSettings, &p.CanManageMessages, &p.CanMentionAll, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get room permission: %w", err)
	}
	return &p, nil
}

func (r *RoomSettingsRepo) UpsertPermissions(ctx context.Context, roomID int64, perms []*model.RoomRolePermission) error {
	for _, p := range perms {
		_, err := r.db.Exec(ctx, `
			INSERT INTO room_role_permissions
				(room_id, role, can_send, can_upload, can_pin, can_manage_members, can_manage_settings, can_manage_messages, can_mention_all)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			ON CONFLICT (room_id, role) DO UPDATE SET
				can_send = EXCLUDED.can_send,
				can_upload = EXCLUDED.can_upload,
				can_pin = EXCLUDED.can_pin,
				can_manage_members = EXCLUDED.can_manage_members,
				can_manage_settings = EXCLUDED.can_manage_settings,
				can_manage_messages = EXCLUDED.can_manage_messages,
				can_mention_all = EXCLUDED.can_mention_all,
				updated_at = NOW()`,
			roomID, p.Role, p.CanSend, p.CanUpload, p.CanPin, p.CanManageMembers, p.CanManageSettings, p.CanManageMessages, p.CanMentionAll,
		)
		if err != nil {
			return fmt.Errorf("upsert room permission: %w", err)
		}
	}
	return nil
}

func (r *RoomSettingsRepo) CreateJoinRequest(ctx context.Context, roomID, userID int64, reason *string) (*model.RoomJoinRequest, error) {
	var jr model.RoomJoinRequest
	err := r.db.QueryRow(ctx, `
		INSERT INTO room_join_requests (room_id, user_id, reason)
		VALUES ($1, $2, $3)
		RETURNING id, room_id, user_id, status, reason, reviewer_id, created_at, updated_at`,
		roomID, userID, reason,
	).Scan(&jr.ID, &jr.RoomID, &jr.UserID, &jr.Status, &jr.Reason, &jr.ReviewerID, &jr.CreatedAt, &jr.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create join request: %w", err)
	}
	return &jr, nil
}

func (r *RoomSettingsRepo) GetJoinRequest(ctx context.Context, id int64) (*model.RoomJoinRequest, error) {
	var jr model.RoomJoinRequest
	err := r.db.QueryRow(ctx, `
		SELECT id, room_id, user_id, status, reason, reviewer_id, created_at, updated_at
		FROM room_join_requests WHERE id = $1`, id,
	).Scan(&jr.ID, &jr.RoomID, &jr.UserID, &jr.Status, &jr.Reason, &jr.ReviewerID, &jr.CreatedAt, &jr.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get join request: %w", err)
	}
	return &jr, nil
}

func (r *RoomSettingsRepo) ListJoinRequests(ctx context.Context, roomID int64) ([]*model.RoomJoinRequest, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, room_id, user_id, status, reason, reviewer_id, created_at, updated_at
		FROM room_join_requests WHERE room_id = $1 ORDER BY created_at DESC`, roomID)
	if err != nil {
		return nil, fmt.Errorf("list join requests: %w", err)
	}
	defer rows.Close()
	var list []*model.RoomJoinRequest
	for rows.Next() {
		var jr model.RoomJoinRequest
		if err := rows.Scan(&jr.ID, &jr.RoomID, &jr.UserID, &jr.Status, &jr.Reason, &jr.ReviewerID, &jr.CreatedAt, &jr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan join request: %w", err)
		}
		list = append(list, &jr)
	}
	return list, nil
}

func (r *RoomSettingsRepo) UpdateJoinRequestStatus(ctx context.Context, id int64, status string, reviewerID int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE room_join_requests SET status = $1, reviewer_id = $2, updated_at = NOW()
		WHERE id = $3`, status, reviewerID, id,
	)
	if err != nil {
		return fmt.Errorf("update join request: %w", err)
	}
	return nil
}

func (r *RoomSettingsRepo) CreateMessageEvent(ctx context.Context, roomID int64, userID *int64, eventType string, meta any) error {
	var raw []byte
	if meta != nil {
		raw, _ = json.Marshal(meta)
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO room_message_events (room_id, user_id, event_type, meta)
		VALUES ($1, $2, $3, $4)`, roomID, userID, eventType, raw,
	)
	if err != nil {
		return fmt.Errorf("create message event: %w", err)
	}
	return nil
}

func (r *RoomSettingsRepo) CreateAuditLog(ctx context.Context, roomID int64, actorID *int64, action string, before, after any) error {
	var beforeRaw []byte
	var afterRaw []byte
	if before != nil {
		beforeRaw, _ = json.Marshal(before)
	}
	if after != nil {
		afterRaw, _ = json.Marshal(after)
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO room_audit_logs (room_id, actor_id, action, before, after)
		VALUES ($1, $2, $3, $4, $5)`, roomID, actorID, action, beforeRaw, afterRaw,
	)
	if err != nil {
		return fmt.Errorf("create room audit log: %w", err)
	}
	return nil
}

type RoomStatsSummary struct {
	TotalMessages int64 `json:"total_messages"`
	ActiveUsers   int64 `json:"active_users"`
}

func (r *RoomSettingsRepo) GetStatsSummary(ctx context.Context, roomID int64, since time.Time) (*RoomStatsSummary, error) {
	var total int64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM messages WHERE room_id = $1 AND created_at >= $2`, roomID, since,
	).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}
	var active int64
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(DISTINCT user_id) FROM messages WHERE room_id = $1 AND created_at >= $2`, roomID, since,
	).Scan(&active)
	if err != nil {
		return nil, fmt.Errorf("count active users: %w", err)
	}
	return &RoomStatsSummary{TotalMessages: total, ActiveUsers: active}, nil
}
