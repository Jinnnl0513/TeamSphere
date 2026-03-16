package model

import "time"

type AuditLog struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"`
	EntityType string   `json:"entity_type"`
	EntityID  int64     `json:"entity_id"`
	Meta      string    `json:"meta"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}
