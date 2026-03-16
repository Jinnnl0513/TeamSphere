package model

import "time"

type LoginAttempt struct {
	Key          string     `json:"key"`
	Attempts     int        `json:"attempts"`
	LockedUntil  *time.Time `json:"locked_until"`
	LastAttemptAt time.Time `json:"last_attempt_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
