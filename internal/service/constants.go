package service

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// ─── Role constants ───

const (
	RoleUser   = "user"
	RoleAdmin  = "admin"
	RoleOwner  = "owner"
	RoleMember = "member"
)

// ─── Message constants ───

const (
	// RecallTimeout is the time window within which a regular user can recall their own message.
	RecallTimeout = recallTimeLimit

	// MaxSearchResults is the maximum number of users returned from a search query.
	MaxSearchResults = 20
)

// ─── Password validation ───

const (
	MinPasswordLength = 8
	MaxPasswordLength = 128
)

// ValidatePassword checks length and complexity requirements.
// Must include upper, lower, and digit.
func ValidatePassword(password string) bool {
	l := len(password)
	if l < MinPasswordLength || l > MaxPasswordLength {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case r >= 'A' && r <= 'Z':
			hasUpper = true
		case r >= 'a' && r <= 'z':
			hasLower = true
		case r >= '0' && r <= '9':
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

// ─── Password utilities ───

// hashPassword hashes a plaintext password using bcrypt.
func hashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

// verifyPassword checks whether a plaintext password matches the bcrypt hash.
// Returns ErrWrongPassword if they don't match.
func verifyPassword(hash, plain string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return ErrWrongPassword
	}
	return nil
}
