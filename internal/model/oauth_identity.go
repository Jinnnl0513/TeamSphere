package model

import "time"

type OAuthIdentity struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Provider  string    `json:"provider"`
	Subject   string    `json:"subject"`
	Email     string    `json:"email,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
