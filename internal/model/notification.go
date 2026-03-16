package model

import "time"

type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	IsRead    bool      `json:"is_read"`
	RefID     int64     `json:"ref_id"`
	CreatedAt time.Time `json:"created_at"`
}
