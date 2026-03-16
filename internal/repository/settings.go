package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type SettingsRepo struct {
	db DBTX
}

func NewSettingsRepo(db DBTX) *SettingsRepo {
	return &SettingsRepo{db: db}
}

// Get retrieves a setting value by key. Returns empty string if not found.
func (r *SettingsRepo) Get(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRow(ctx,
		`SELECT value FROM system_settings WHERE key = $1`, key,
	).Scan(&value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return value, nil
}

// Set upserts a setting key-value pair.
func (r *SettingsRepo) Set(ctx context.Context, key, value string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO system_settings (key, value, updated_at)
		 VALUES ($1, $2, NOW())
		 ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

// GetAll returns all system settings as a key→value map.
func (r *SettingsRepo) GetAll(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.Query(ctx, `SELECT key, value FROM system_settings ORDER BY key`)
	if err != nil {
		return nil, fmt.Errorf("get all settings: %w", err)
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		m[k] = v
	}
	return m, rows.Err()
}

// GetByPrefix returns settings whose key starts with the given prefix.
func (r *SettingsRepo) GetByPrefix(ctx context.Context, prefix string) (map[string]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT key, value FROM system_settings WHERE key LIKE $1 ORDER BY key`,
		prefix+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("get settings by prefix: %w", err)
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		m[k] = v
	}
	return m, rows.Err()
}
