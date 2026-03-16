package model

import "time"

type RoomSettings struct {
	ID                 int64      `json:"id"`
	RoomID             int64      `json:"room_id"`
	IsPublic           bool       `json:"is_public"`
	RequireApproval    bool       `json:"require_approval"`
	ReadOnly           bool       `json:"read_only"`
	Archived           bool       `json:"archived"`
	Topic              *string    `json:"topic,omitempty"`
	AvatarURL          *string    `json:"avatar_url,omitempty"`
	SlowModeSeconds    int        `json:"slow_mode_seconds"`
	MessageRetention   int        `json:"message_retention_days"`
	ContentFilterMode  string     `json:"content_filter_mode"`
	BlockedKeywords    []string   `json:"blocked_keywords,omitempty"`
	AllowedLinkDomains []string   `json:"allowed_link_domains,omitempty"`
	BlockedLinkDomains []string   `json:"blocked_link_domains,omitempty"`
	AllowedFileTypes   []string   `json:"allowed_file_types,omitempty"`
	MaxFileSizeMB      int        `json:"max_file_size_mb"`
	PinLimit           int        `json:"pin_limit"`
	NotifyMode         string     `json:"notify_mode"`
	NotifyKeywords     []string   `json:"notify_keywords,omitempty"`
	DNDStart           *string    `json:"dnd_start,omitempty"`
	DNDEnd             *string    `json:"dnd_end,omitempty"`
	AntiSpamRate       int        `json:"anti_spam_rate"`
	AntiSpamWindowSec  int        `json:"anti_spam_window_sec"`
	AntiRepeat         bool       `json:"anti_repeat"`
	StatsEnabled       bool       `json:"stats_enabled"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type RoomRolePermission struct {
	ID               int64     `json:"id"`
	RoomID           int64     `json:"room_id"`
	Role             string    `json:"role"`
	CanSend          bool      `json:"can_send"`
	CanUpload        bool      `json:"can_upload"`
	CanPin           bool      `json:"can_pin"`
	CanManageMembers bool      `json:"can_manage_members"`
	CanManageSettings bool     `json:"can_manage_settings"`
	CanManageMessages bool     `json:"can_manage_messages"`
	CanMentionAll    bool      `json:"can_mention_all"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RoomJoinRequest struct {
	ID         int64      `json:"id"`
	RoomID     int64      `json:"room_id"`
	UserID     int64      `json:"user_id"`
	Status     string     `json:"status"`
	Reason     *string    `json:"reason,omitempty"`
	ReviewerID *int64     `json:"reviewer_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type RoomMessageEvent struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	UserID    *int64    `json:"user_id,omitempty"`
	EventType string    `json:"event_type"`
	Meta      any       `json:"meta,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type RoomAuditLog struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	ActorID   *int64    `json:"actor_id,omitempty"`
	Action    string    `json:"action"`
	Before    any       `json:"before,omitempty"`
	After     any       `json:"after,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
