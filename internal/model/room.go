package model

import "time"

type Room struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   *int64    `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type RoomMember struct {
	RoomID    int64      `json:"room_id"`
	UserID    int64      `json:"user_id"`
	Role      string     `json:"role"`
	MutedUntil *time.Time `json:"muted_until,omitempty"`
	JoinedAt  time.Time  `json:"joined_at"`
}

type RoomInvite struct {
	ID        int64     `json:"id"`
	RoomID    int64     `json:"room_id"`
	InviterID int64     `json:"inviter_id"`
	InviteeID int64     `json:"invitee_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
