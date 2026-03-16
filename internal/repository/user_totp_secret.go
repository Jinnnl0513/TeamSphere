package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

type UserTOTPSecretRepo struct {
	db DBTX
}

func NewUserTOTPSecretRepo(db DBTX) *UserTOTPSecretRepo {
	return &UserTOTPSecretRepo{db: db}
}

func (r *UserTOTPSecretRepo) GetByUserID(ctx context.Context, userID int64) (*model.UserTOTPSecret, error) {
	var s model.UserTOTPSecret
	err := r.db.QueryRow(ctx,
		`SELECT user_id, secret_enc, enabled, created_at, updated_at, last_used_at
		 FROM user_totp_secrets WHERE user_id = $1`, userID,
	).Scan(&s.UserID, &s.SecretEnc, &s.Enabled, &s.CreatedAt, &s.UpdatedAt, &s.LastUsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user totp secret: %w", err)
	}
	return &s, nil
}

func (r *UserTOTPSecretRepo) Upsert(ctx context.Context, userID int64, secretEnc string, enabled bool) (*model.UserTOTPSecret, error) {
	var s model.UserTOTPSecret
	err := r.db.QueryRow(ctx,
		`INSERT INTO user_totp_secrets (user_id, secret_enc, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, NOW(), NOW())
		 ON CONFLICT (user_id) DO UPDATE
		 SET secret_enc = EXCLUDED.secret_enc,
		     enabled = EXCLUDED.enabled,
		     updated_at = NOW()
		 RETURNING user_id, secret_enc, enabled, created_at, updated_at, last_used_at`,
		userID, secretEnc, enabled,
	).Scan(&s.UserID, &s.SecretEnc, &s.Enabled, &s.CreatedAt, &s.UpdatedAt, &s.LastUsedAt)
	if err != nil {
		return nil, fmt.Errorf("upsert user totp secret: %w", err)
	}
	return &s, nil
}

func (r *UserTOTPSecretRepo) SetEnabled(ctx context.Context, userID int64, enabled bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE user_totp_secrets SET enabled = $1, updated_at = NOW() WHERE user_id = $2`,
		enabled, userID,
	)
	if err != nil {
		return fmt.Errorf("set user totp enabled: %w", err)
	}
	return nil
}

func (r *UserTOTPSecretRepo) UpdateLastUsed(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE user_totp_secrets SET last_used_at = NOW(), updated_at = NOW() WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("update user totp last used: %w", err)
	}
	return nil
}

func (r *UserTOTPSecretRepo) Delete(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM user_totp_secrets WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete user totp secret: %w", err)
	}
	return nil
}
