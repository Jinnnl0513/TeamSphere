package model

import "time"

type User struct {
	ID              int64      `json:"id"`
	Username        string     `json:"username"`
	Password        string     `json:"-"`
	AvatarURL       string     `json:"avatar_url"`
	Bio             string     `json:"bio"`
	ProfileColor    string     `json:"profile_color"`
	Email           string     `json:"email"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	Role            string     `json:"role"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UserInfo is the public-facing user object embedded in WS messages and API responses.
// NOTE: Email is intentionally excluded — it must not be broadcast publicly.
// Email is only returned for the user themselves via /users/me.
type UserInfo struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	AvatarURL    string `json:"avatar_url"`
	Bio          string `json:"bio"`
	ProfileColor string `json:"profile_color"`
}

func (u *User) ToInfo() UserInfo {
	return UserInfo{
		ID:           u.ID,
		Username:     u.Username,
		AvatarURL:    u.AvatarURL,
		Bio:          u.Bio,
		ProfileColor: u.ProfileColor,
	}
}

type TokenBlacklist struct {
	ID        int64     `json:"id"`
	TokenJTI  string    `json:"token_jti"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// EmailVerification represents a pending email verification code.
type EmailVerification struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	Attempts  int       `json:"attempts"`
	Used      bool      `json:"used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// RefreshToken represents a long-lived token used to obtain new access tokens.
type RefreshToken struct {
	ID         int64      `json:"id"`
	UserID     int64      `json:"user_id"`
	TokenHash  string     `json:"token_hash"`
	ExpiresAt  time.Time  `json:"expires_at"`
	CreatedAt  time.Time  `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	IPAddress  *string    `json:"ip_address,omitempty"`
	UserAgent  *string    `json:"user_agent,omitempty"`
	DeviceName *string    `json:"device_name,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}
