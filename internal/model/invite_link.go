package model

import "time"

// InviteLink represents a short-code based invite link for a room.
type InviteLink struct {
	ID              int64      `json:"id"`
	Code            string     `json:"code"`
	RoomID          int64      `json:"room_id"`
	CreatorID       int64      `json:"creator_id"`
	CreatorName string `json:"creator_name"` // filled via JOIN, not a DB column
	MaxUses         int        `json:"max_uses"`         // 0 = unlimited
	Uses            int        `json:"uses"`
	ExpiresAt       *time.Time `json:"expires_at"` // nil = never
	CreatedAt       time.Time  `json:"created_at"`
}
