package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

type OAuthIdentityRepo struct {
	db DBTX
}

func NewOAuthIdentityRepo(db DBTX) *OAuthIdentityRepo {
	return &OAuthIdentityRepo{db: db}
}

func (r *OAuthIdentityRepo) GetByProviderSubject(ctx context.Context, provider, subject string) (*model.OAuthIdentity, error) {
	row := &model.OAuthIdentity{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, provider, subject, email, created_at
		 FROM oauth_identities WHERE provider = $1 AND subject = $2`,
		provider, subject,
	).Scan(&row.ID, &row.UserID, &row.Provider, &row.Subject, &row.Email, &row.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get oauth identity: %w", err)
	}
	return row, nil
}

func (r *OAuthIdentityRepo) GetByUserProvider(ctx context.Context, userID int64, provider string) (*model.OAuthIdentity, error) {
	row := &model.OAuthIdentity{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, provider, subject, email, created_at
		 FROM oauth_identities WHERE user_id = $1 AND provider = $2`,
		userID, provider,
	).Scan(&row.ID, &row.UserID, &row.Provider, &row.Subject, &row.Email, &row.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get oauth identity by user: %w", err)
	}
	return row, nil
}

func (r *OAuthIdentityRepo) Create(ctx context.Context, identity *model.OAuthIdentity) (*model.OAuthIdentity, error) {
	row := &model.OAuthIdentity{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO oauth_identities (user_id, provider, subject, email)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, provider, subject, email, created_at`,
		identity.UserID, identity.Provider, identity.Subject, identity.Email,
	).Scan(&row.ID, &row.UserID, &row.Provider, &row.Subject, &row.Email, &row.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create oauth identity: %w", err)
	}
	return row, nil
}
