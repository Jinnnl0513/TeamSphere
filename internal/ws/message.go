package ws

import (
	"encoding/json"
	"time"

	"github.com/teamsphere/server/internal/model"
)

// ─── Envelope ───

// Envelope is the top-level WS message structure.
type Envelope struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// NewEnvelope creates an envelope with the given type and data payload.
func NewEnvelope(msgType string, data any) (*Envelope, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return &Envelope{Type: msgType, Data: raw}, nil
}

// MustEnvelope is like NewEnvelope but panics on marshal error (use for static data).
func MustEnvelope(msgType string, data any) *Envelope {
	e, err := NewEnvelope(msgType, data)
	if err != nil {
		panic(err)
	}
	return e
}

// ─── Client �?Server message types ───

const (
	TypeChat      = "chat"
	TypeJoinRoom  = "join_room"
	TypeLeaveRoom = "leave_room"
	TypeTyping    = "typing"
	TypeDM        = "dm"
	TypeDMTyping  = "dm_typing"
)

// ChatMessage is the client→server chat payload.
type ChatMessage struct {
	RoomID      int64   `json:"room_id"`
	Content     string  `json:"content"`
	ClientMsgID string  `json:"client_msg_id"`
	MsgType     string  `json:"msg_type,omitempty"` // "text" (default) or "image"
	FileSize    *int64  `json:"file_size,omitempty"`
	MimeType    *string `json:"mime_type,omitempty"`
	ReplyToID   *int64  `json:"reply_to_id,omitempty"`
}

// JoinRoomMessage is the client→server join_room payload.
type JoinRoomMessage struct {
	RoomID int64 `json:"room_id"`
}

// LeaveRoomMessage is the client→server leave_room payload.
type LeaveRoomMessage struct {
	RoomID int64 `json:"room_id"`
}

// TypingMessage is the client→server typing payload.
type TypingMessage struct {
	RoomID int64 `json:"room_id"`
}

// DMMessage is the client→server dm payload.
type DMMessage struct {
	TargetUserID int64   `json:"target_user_id"`
	Content      string  `json:"content"`
	ClientMsgID  string  `json:"client_msg_id"`
	MsgType      string  `json:"msg_type,omitempty"`
	FileSize     *int64  `json:"file_size,omitempty"`
	MimeType     *string `json:"mime_type,omitempty"`
	ReplyToID    *int64  `json:"reply_to_id,omitempty"`
}

// DMTypingMessage is the client→server dm_typing payload (Batch 6).
type DMTypingMessage struct {
	TargetUserID int64 `json:"target_user_id"`
}

// ─── Server �?Client message types ───

const (
	TypeChatAck      = "chat_ack"
	TypeDMSent       = "dm_sent"
	TypeSystem       = "system"
	TypeError        = "error"
	TypeOnlineUsers  = "online_users"
	TypeMsgRecalled  = "msg_recalled"
	TypeDMRecalled   = "dm_recalled"
	TypeMsgEdited    = "msg_edited"
	TypeDMEdited     = "dm_edited"
	TypeThreadReply  = "thread_reply"
	TypeMsgPinned    = "msg_pinned"
	TypeMsgUnpinned  = "msg_unpinned"
	TypeUnreadUpdate = "unread_update"
	TypeDMRead       = "dm_read"

	// Room management notifications
	TypeRoomInvite            = "room_invite"
	TypeRoomInviteAccepted    = "room_invite_accepted"
	TypeRoomInviteDeclined    = "room_invite_declined"
	TypeRoomMemberJoined      = "room_member_joined"
	TypeRoomMemberLeft        = "room_member_left"
	TypeRoomMemberKicked      = "room_member_kicked"
	TypeRoomMemberMuted       = "room_member_muted"
	TypeRoomMemberUnmuted     = "room_member_unmuted"
	TypeRoomMemberRoleChanged = "room_member_role_changed"
	TypeRoomUpdated           = "room_updated"
	TypeRoomDeleted           = "room_deleted"
	TypeRoomOwnerTransferred  = "room_owner_transferred"

	// Friend notifications (Batch 6)
	TypeFriendRequest     = "friend_request"
	TypeFriendAccepted    = "friend_accepted"
	TypeFriendRejected    = "friend_rejected"
	TypeFriendRemoved     = "friend_removed"
	TypeFriendOnline      = "friend_online"
	TypeFriendOffline     = "friend_offline"
	TypeFriendsOnlineList = "friends_online_list"

	// Mentioned notification (Batch 7)
	TypeMentioned = "mentioned"

	// Reaction notifications (Phase 3)
	TypeReactionAdded   = "reaction_added"
	TypeReactionRemoved = "reaction_removed"
)

// ChatBroadcast is the server→client chat payload.
type ChatBroadcast struct {
	ID          int64              `json:"id"`
	ClientMsgID string             `json:"client_msg_id,omitempty"`
	Content     string             `json:"content"`
	MsgType     string             `json:"msg_type"`
	FileSize    *int64             `json:"file_size,omitempty"`
	MimeType    *string            `json:"mime_type,omitempty"`
	ForwardMeta *model.ForwardInfo `json:"forward_meta,omitempty"`
	User        model.UserInfo     `json:"user"`
	RoomID      int64              `json:"room_id"`
	Mentions    []int64            `json:"mentions"`
	ReplyTo     *model.ReplyInfo   `json:"reply_to,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

// ChatAck is the server→sender confirmation after a room message is persisted.
type ChatAck struct {
	ClientMsgID string    `json:"client_msg_id"`
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
}

// SystemMessage is the server→client system notification.
type SystemMessage struct {
	Content string `json:"content"`
	RoomID  int64  `json:"room_id"`
}

// ErrorMessage is the server→client error payload.
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OnlineUsersMessage lists online members for a room.
type OnlineUsersMessage struct {
	RoomID int64            `json:"room_id"`
	Users  []model.UserInfo `json:"users"`
}

// TypingBroadcast is the server→client typing indicator.
type TypingBroadcast struct {
	RoomID   int64  `json:"room_id"`
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// DMBroadcast is the server→client dm payload.
type DMBroadcast struct {
	ID          int64              `json:"id"`
	ClientMsgID string             `json:"client_msg_id,omitempty"`
	Content     string             `json:"content"`
	MsgType     string             `json:"msg_type"`
	FileSize    *int64             `json:"file_size,omitempty"`
	MimeType    *string            `json:"mime_type,omitempty"`
	ForwardMeta *model.ForwardInfo `json:"forward_meta,omitempty"`
	User        model.UserInfo     `json:"user"`
	ReplyTo     *model.ReplyInfo   `json:"reply_to,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
}

// DMSent is the server→sender confirmation after a DM is persisted.
type DMSent struct {
	ClientMsgID string    `json:"client_msg_id"`
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	Delivered   bool      `json:"delivered"`
}

// DMTypingBroadcast is the server→client dm_typing indicator.
type DMTypingBroadcast struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
}

// ─── WS Error Codes ───

const (
	ErrCodeInvalidFormat    = "INVALID_FORMAT"
	ErrCodeUnknownType      = "UNKNOWN_TYPE"
	ErrCodeNotRoomMember    = "NOT_ROOM_MEMBER"
	ErrCodeRoomNotFound     = "ROOM_NOT_FOUND"
	ErrCodeMessageTooLong   = "MESSAGE_TOO_LONG"
	ErrCodeRateLimited      = "RATE_LIMITED"
	ErrCodeNotFriends       = "NOT_FRIENDS"
	ErrCodeMuted            = "MUTED"
	ErrCodeMentionAllForbid = "MENTION_ALL_FORBIDDEN"
	ErrCodeRoomReadOnly     = "ROOM_READ_ONLY"
	ErrCodeSendForbidden    = "SEND_FORBIDDEN"
	ErrCodeContentBlocked   = "CONTENT_BLOCKED"
	ErrCodeSlowMode         = "SLOW_MODE"
	ErrCodeSpamBlocked      = "SPAM_BLOCKED"
	ErrCodeUploadForbidden  = "UPLOAD_FORBIDDEN"
	ErrCodeLinkBlocked      = "LINK_BLOCKED"
)

// ─── Action (REST �?Hub broadcast) ───

// Action represents a command from the REST/Service layer to the Hub
// for broadcasting notifications to connected clients.
type Action struct {
	Type   string // action type identifier
	RoomID int64  // target room (0 = not room-scoped)
	UserID int64  // target user (0 = broadcast to room)
	Data   any    // payload to marshal and send
}

// ReactionBroadcast is the server→client reaction change payload.
type ReactionBroadcast struct {
	MessageID   int64   `json:"message_id"`
	MessageType string  `json:"message_type"` // "room" | "dm"
	Emoji       string  `json:"emoji"`
	UserID      int64   `json:"user_id"`
	Count       int     `json:"count"`
	UserIDs     []int64 `json:"user_ids"` // updated full list of reactor IDs (capped at 20)
}
