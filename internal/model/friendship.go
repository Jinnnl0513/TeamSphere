package model

import "time"

type Friendship struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	FriendID  int64     `json:"friend_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SystemSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
