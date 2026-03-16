package model

import "time"

type MessageRead struct {
	UserID        int64     `json:"user_id"`
	RoomID        int64     `json:"room_id"`
	LastReadMsgID int64     `json:"last_read_msg_id"`
	ReadAt        time.Time `json:"read_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
