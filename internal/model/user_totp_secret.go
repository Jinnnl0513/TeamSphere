package model

import "time"

type UserTOTPSecret struct {
	UserID     int64      `json:"user_id"`
	SecretEnc  string     `json:"-"`
	Enabled    bool       `json:"enabled"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}
