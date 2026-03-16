package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

type LoginAttemptRepo struct {
	db DBTX
}

func NewLoginAttemptRepo(db DBTX) *LoginAttemptRepo {
	return &LoginAttemptRepo{db: db}
}

func (r *LoginAttemptRepo) Get(ctx context.Context, key string) (*model.LoginAttempt, error) {
	la := &model.LoginAttempt{}
	err := r.db.QueryRow(ctx, `
		SELECT key, attempts, locked_until, last_attempt_at, created_at, updated_at
		FROM login_attempts WHERE key = $1
	`, key).Scan(&la.Key, &la.Attempts, &la.LockedUntil, &la.LastAttemptAt, &la.CreatedAt, &la.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get login attempt: %w", err)
	}
	return la, nil
}

func (r *LoginAttemptRepo) Upsert(ctx context.Context, key string, attempts int, lockedUntil *time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO login_attempts (key, attempts, locked_until, last_attempt_at, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW(), NOW())
		ON CONFLICT (key)
		DO UPDATE SET attempts = EXCLUDED.attempts,
		              locked_until = EXCLUDED.locked_until,
		              last_attempt_at = NOW(),
		              updated_at = NOW()
	`, key, attempts, lockedUntil)
	if err != nil {
		return fmt.Errorf("upsert login attempt: %w", err)
	}
	return nil
}

func (r *LoginAttemptRepo) Reset(ctx context.Context, key string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE login_attempts
		SET attempts = 0, locked_until = NULL, updated_at = NOW()
		WHERE key = $1
	`, key)
	if err != nil {
		return fmt.Errorf("reset login attempt: %w", err)
	}
	return nil
}

