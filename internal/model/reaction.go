package model

import "time"

// Reaction represents a single emoji reaction on a message.
type Reaction struct {
	ID          int64     `json:"id"`
	MessageID   int64     `json:"message_id"`
	MessageType string    `json:"message_type"` // "room" | "dm"
	UserID      int64     `json:"user_id"`
	Emoji       string    `json:"emoji"`
	CreatedAt   time.Time `json:"created_at"`
}

// ReactionSummary aggregates reactions for a message (for API responses).
type ReactionSummary struct {
	Emoji   string  `json:"emoji"`
	Count   int     `json:"count"`
	UserIDs []int64 `json:"user_ids"` // list of reactors, capped at 20
}
