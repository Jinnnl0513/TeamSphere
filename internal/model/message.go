package model

import "time"

type Message struct {
	ID          int64      `json:"id"`
	Content     string     `json:"content"`
	UserID      *int64     `json:"user_id"`
	RoomID      int64      `json:"room_id"`
	MsgType     string     `json:"msg_type"`
	Mentions    []int64    `json:"mentions"`
	FileSize   *int64     `json:"file_size,omitempty"`
	MimeType   *string    `json:"mime_type,omitempty"`
	ReplyToID   *int64     `json:"reply_to_id,omitempty"`
	ForwardMeta *ForwardInfo `json:"forward_meta,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	ClientMsgID *string    `json:"client_msg_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type DirectMessage struct {
	ID          int64      `json:"id"`
	Content     string     `json:"content"`
	SenderID    int64      `json:"sender_id"`
	ReceiverID  int64      `json:"receiver_id"`
	MsgType     string     `json:"msg_type"`
	FileSize   *int64     `json:"file_size,omitempty"`
	MimeType   *string    `json:"mime_type,omitempty"`
	ReplyToID   *int64     `json:"reply_to_id,omitempty"`
	ForwardMeta *ForwardInfo `json:"forward_meta,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	ClientMsgID *string    `json:"client_msg_id,omitempty"`
	ReadAt      *time.Time `json:"read_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ForwardInfo is a snapshot of forwarded message metadata.
type ForwardInfo struct {
	Type      string   `json:"type"` // "room" | "dm"
	ID        int64    `json:"id"`
	MsgType   string   `json:"msg_type"`
	Content   string   `json:"content"`
	IsDeleted bool     `json:"is_deleted"`
	User      UserInfo `json:"user"`
}

// ReplyInfo is a trimmed snapshot of the quoted message included in broadcast payloads.
// Content is truncated to 200 runes. When the quoted message is recalled, Content
// is set to the empty string and IsDeleted is true.
type ReplyInfo struct {
	ID        int64    `json:"id"`
	Content   string   `json:"content"` // truncated to 200 runes; empty when recalled
	MsgType   string   `json:"msg_type"`
	IsDeleted bool     `json:"is_deleted"` // true when the quoted message was recalled
	User      UserInfo `json:"user"`
}
