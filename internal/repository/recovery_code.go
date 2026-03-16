package repository

import (
	"context"
	"fmt"
)

type RecoveryCodeRepo struct {
	db DBTX
}

func NewRecoveryCodeRepo(db DBTX) *RecoveryCodeRepo {
	return &RecoveryCodeRepo{db: db}
}

func (r *RecoveryCodeRepo) ReplaceCodes(ctx context.Context, userID int64, codeHashes []string) error {
	if len(codeHashes) == 0 {
		return nil
	}
	return withTx(ctx, r.db, func(tx DBTX) error {
		if _, err := tx.Exec(ctx, `DELETE FROM user_recovery_codes WHERE user_id = $1`, userID); err != nil {
			return fmt.Errorf("delete recovery codes: %w", err)
		}
		for _, hash := range codeHashes {
			if _, err := tx.Exec(ctx,
				`INSERT INTO user_recovery_codes (user_id, code_hash) VALUES ($1, $2)`,
				userID, hash,
			); err != nil {
				return fmt.Errorf("insert recovery code: %w", err)
			}
		}
		return nil
	})
}

func (r *RecoveryCodeRepo) ConsumeCode(ctx context.Context, userID int64, codeHash string) (bool, error) {
	var updated bool
	err := r.db.QueryRow(ctx,
		`UPDATE user_recovery_codes
		 SET used_at = NOW()
		 WHERE user_id = $1 AND code_hash = $2 AND used_at IS NULL
		 RETURNING true`,
		userID, codeHash,
	).Scan(&updated)
	if err != nil {
		return false, nil
	}
	return updated, nil
}

func (r *RecoveryCodeRepo) CountAvailable(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM user_recovery_codes WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count recovery codes: %w", err)
	}
	return count, nil
}
