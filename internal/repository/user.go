package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepo struct {
	db DBTX
}

func NewUserRepo(db DBTX) *UserRepo {
	return &UserRepo{db: db}
}

// scanUser scans all user fields from a row.
func scanUser(row pgx.Row, u *model.User) error {
	return row.Scan(
		&u.ID, &u.Username, &u.Password,
		&u.AvatarURL, &u.Bio, &u.ProfileColor,
		&u.Email, &u.EmailVerifiedAt,
		&u.Role, &u.DeletedAt, &u.CreatedAt, &u.UpdatedAt,
	)
}

const selectUserCols = `id, username, password, avatar_url, bio, profile_color, email, email_verified_at, role, deleted_at, created_at, updated_at`

// Create inserts a new user and returns it with the generated ID.
func (r *UserRepo) Create(ctx context.Context, username, password, role, email string) (*model.User, error) {
	u := &model.User{}
	var err error
	err = scanUser(r.db.QueryRow(ctx,
		`INSERT INTO users (username, password, role, email)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+selectUserCols,
		username, password, role, email,
	), u)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}

// GetByID retrieves a user by ID.
func (r *UserRepo) GetByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+selectUserCols+` FROM users WHERE id = $1 AND deleted_at IS NULL`, id,
	), u)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// ListUserInfosByIDs returns basic user info for a list of IDs.
func (r *UserRepo) ListUserInfosByIDs(ctx context.Context, ids []int64) ([]model.UserInfo, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, username, avatar_url FROM users WHERE id = ANY($1) AND deleted_at IS NULL`,
		pgtype.FlatArray[int64](ids),
	)
	if err != nil {
		return nil, fmt.Errorf("list user infos: %w", err)
	}
	defer rows.Close()

	var out []model.UserInfo
	for rows.Next() {
		var info model.UserInfo
		if err := rows.Scan(&info.ID, &info.Username, &info.AvatarURL); err != nil {
			return nil, fmt.Errorf("scan user info: %w", err)
		}
		out = append(out, info)
	}
	return out, rows.Err()
}

// GetByUsername retrieves a user by username.
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+selectUserCols+` FROM users WHERE username = $1 AND deleted_at IS NULL`, username,
	), u)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

// ExistsAny returns true if there is at least one user in the users table.
func (r *UserRepo) ExistsAny(ctx context.Context) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE deleted_at IS NULL)`).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check users exist: %w", err)
	}
	return exists, nil
}

// BlacklistToken inserts a token JTI into the blacklist.
func (r *UserRepo) BlacklistToken(ctx context.Context, entry *model.TokenBlacklist) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO token_blacklist (token_jti, user_id, expires_at)
		 VALUES ($1, $2, $3)`,
		entry.TokenJTI, entry.UserID, entry.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("blacklist token: %w", err)
	}
	return nil
}

// IsTokenBlacklisted checks whether a JTI is in the blacklist.
func (r *UserRepo) IsTokenBlacklisted(ctx context.Context, jti string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM token_blacklist WHERE token_jti = $1)`, jti,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check token blacklist: %w", err)
	}
	return exists, nil
}

// CleanExpiredTokens removes expired entries from the blacklist.
func (r *UserRepo) CleanExpiredTokens(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM token_blacklist WHERE expires_at < NOW()`,
	)
	if err != nil {
		return 0, fmt.Errorf("clean expired tokens: %w", err)
	}
	return tag.RowsAffected(), nil
}

// UpdateBioAndColor updates a user's bio and profile_color.
func (r *UserRepo) UpdateBioAndColor(ctx context.Context, id int64, bio, profileColor string) (*model.User, error) {
	u := &model.User{}
	err := scanUser(r.db.QueryRow(ctx,
		`UPDATE users SET bio = $1, profile_color = $2, updated_at = NOW() WHERE id = $3
		 RETURNING `+selectUserCols,
		bio, profileColor, id,
	), u)
	if err != nil {
		return nil, fmt.Errorf("update bio and color: %w", err)
	}
	return u, nil
}

// UpdatePassword updates a user's password hash.
func (r *UserRepo) UpdatePassword(ctx context.Context, id int64, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2`,
		passwordHash, id,
	)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// UpdateAvatar updates a user's avatar URL.
func (r *UserRepo) UpdateAvatar(ctx context.Context, id int64, avatarURL string) (*model.User, error) {
	u := &model.User{}
	err := scanUser(r.db.QueryRow(ctx,
		`UPDATE users SET avatar_url = $1, updated_at = NOW() WHERE id = $2
		 RETURNING `+selectUserCols,
		avatarURL, id,
	), u)
	if err != nil {
		return nil, fmt.Errorf("update avatar: %w", err)
	}
	return u, nil
}

// SoftDelete marks a user as deleted.
func (r *UserRepo) SoftDelete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET deleted_at = NOW(), updated_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	return nil
}

// ─── Admin queries ───

// ListAll returns paginated users. Includes soft-deleted users.
func (r *UserRepo) ListAll(ctx context.Context, offset, limit int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := r.db.Query(ctx,
		`SELECT `+selectUserCols+` FROM users ORDER BY id ASC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list all users: %w", err)
	}
	defer rows.Close()

	var users []*model.User
	for rows.Next() {
		u := &model.User{}
		if err := scanUser(rows, u); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

// CountAll returns the total number of users (including soft-deleted).
func (r *UserRepo) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count all users: %w", err)
	}
	return count, nil
}

// CountActive returns the number of non-deleted users.
func (r *UserRepo) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count active users: %w", err)
	}
	return count, nil
}

// UpdateRole changes a user's system role.
func (r *UserRepo) UpdateRole(ctx context.Context, id int64, role string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`,
		role, id,
	)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

// HardDelete physically removes a user from the database.
func (r *UserRepo) HardDelete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("hard delete user: %w", err)
	}
	return nil
}

// GetByEmail retrieves a user by email. Returns nil if not found.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	u := &model.User{}
	err := scanUser(r.db.QueryRow(ctx,
		`SELECT `+selectUserCols+` FROM users WHERE email = $1 AND email != '' AND deleted_at IS NULL`, email,
	), u)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return u, nil
}

// SetEmailVerified marks a user's email as verified.
func (r *UserRepo) SetEmailVerified(ctx context.Context, id int64, email string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET email = $1, email_verified_at = NOW(), updated_at = NOW() WHERE id = $2`,
		email, id,
	)
	if err != nil {
		return fmt.Errorf("set email verified: %w", err)
	}
	return nil
}

// ─── Email verification codes ───

// CreateEmailVerification inserts a new verification code.
func (r *UserRepo) CreateEmailVerification(ctx context.Context, email, code string, expiresAt time.Time) (*model.EmailVerification, error) {
	v := &model.EmailVerification{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO email_verifications (email, code, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, email, code, attempts, used, expires_at, created_at`,
		email, code, expiresAt,
	).Scan(&v.ID, &v.Email, &v.Code, &v.Attempts, &v.Used, &v.ExpiresAt, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create email verification: %w", err)
	}
	return v, nil
}

// GetLatestVerification returns the latest unused, non-expired verification for an email.
func (r *UserRepo) GetLatestVerification(ctx context.Context, email string) (*model.EmailVerification, error) {
	v := &model.EmailVerification{}
	err := r.db.QueryRow(ctx,
		`SELECT id, email, code, attempts, used, expires_at, created_at
		 FROM email_verifications
		 WHERE email = $1 AND used = false AND expires_at > NOW()
		 ORDER BY created_at DESC LIMIT 1`,
		email,
	).Scan(&v.ID, &v.Email, &v.Code, &v.Attempts, &v.Used, &v.ExpiresAt, &v.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest verification: %w", err)
	}
	return v, nil
}

// MarkVerificationUsed marks a verification code as used.
func (r *UserRepo) MarkVerificationUsed(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE email_verifications SET used = true WHERE id = $1`, id,
	)
	if err != nil {
		return fmt.Errorf("mark verification used: %w", err)
	}
	return nil
}

// IncrementVerificationAttempts increments the attempt counter.
func (r *UserRepo) IncrementVerificationAttempts(ctx context.Context, id int64) (int, error) {
	var attempts int
	err := r.db.QueryRow(ctx,
		`UPDATE email_verifications SET attempts = attempts + 1 WHERE id = $1 RETURNING attempts`, id,
	).Scan(&attempts)
	if err != nil {
		return 0, fmt.Errorf("increment verification attempts: %w", err)
	}
	return attempts, nil
}

// CleanExpiredVerifications removes expired verification codes.
func (r *UserRepo) CleanExpiredVerifications(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM email_verifications WHERE expires_at < NOW()`,
	)
	if err != nil {
		return 0, fmt.Errorf("clean expired verifications: %w", err)
	}
	return tag.RowsAffected(), nil
}

// ─── Refresh Token operations ───

// CreateRefreshToken stores a new refresh token hash.
func (r *UserRepo) CreateRefreshToken(ctx context.Context, userID int64, tokenHash string, expiresAt time.Time, ipAddress, userAgent, deviceName *string) (*model.RefreshToken, error) {
	rt := &model.RefreshToken{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at, ip_address, user_agent, device_name, last_used_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())
		 RETURNING id, user_id, token_hash, expires_at, created_at, revoked_at, ip_address, user_agent, device_name, last_used_at`,
		userID, tokenHash, expiresAt, ipAddress, userAgent, deviceName,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt, &rt.IPAddress, &rt.UserAgent, &rt.DeviceName, &rt.LastUsedAt)
	if err != nil {
		return nil, fmt.Errorf("create refresh token: %w", err)
	}
	return rt, nil
}

// GetRefreshTokenByHash retrieves a refresh token by its hash.
func (r *UserRepo) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error) {
	rt := &model.RefreshToken{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked_at, ip_address, user_agent, device_name, last_used_at
		 FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt, &rt.IPAddress, &rt.UserAgent, &rt.DeviceName, &rt.LastUsedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get refresh token: %w", err)
	}
	return rt, nil
}

// RevokeRefreshToken marks a refresh token as revoked.
func (r *UserRepo) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

// RevokeAllRefreshTokensForUser revokes all refresh tokens for a user.
func (r *UserRepo) RevokeAllRefreshTokensForUser(ctx context.Context, userID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("revoke all refresh tokens: %w", err)
	}
	return nil
}

// RevokeRefreshTokenByID revokes a refresh token for a user by ID.
func (r *UserRepo) RevokeRefreshTokenByID(ctx context.Context, userID, tokenID int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND id = $2 AND revoked_at IS NULL`,
		userID, tokenID,
	)
	if err != nil {
		return fmt.Errorf("revoke refresh token by id: %w", err)
	}
	return nil
}

// ListRefreshTokens returns all refresh tokens for a user.
func (r *UserRepo) ListRefreshTokens(ctx context.Context, userID int64) ([]*model.RefreshToken, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked_at, ip_address, user_agent, device_name, last_used_at
		 FROM refresh_tokens WHERE user_id = $1 ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list refresh tokens: %w", err)
	}
	defer rows.Close()

	var items []*model.RefreshToken
	for rows.Next() {
		rt := &model.RefreshToken{}
		if err := rows.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.RevokedAt, &rt.IPAddress, &rt.UserAgent, &rt.DeviceName, &rt.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan refresh token: %w", err)
		}
		items = append(items, rt)
	}
	return items, rows.Err()
}

// UpdateRefreshTokenLastUsed updates last_used_at for a refresh token.
func (r *UserRepo) UpdateRefreshTokenLastUsed(ctx context.Context, tokenHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET last_used_at = NOW() WHERE token_hash = $1`,
		tokenHash,
	)
	if err != nil {
		return fmt.Errorf("update refresh token last used: %w", err)
	}
	return nil
}

// RevokeOtherRefreshTokens revokes all refresh tokens except the current one.
func (r *UserRepo) RevokeOtherRefreshTokens(ctx context.Context, userID int64, tokenHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = NOW()
		 WHERE user_id = $1 AND token_hash <> $2 AND revoked_at IS NULL`,
		userID, tokenHash,
	)
	if err != nil {
		return fmt.Errorf("revoke other refresh tokens: %w", err)
	}
	return nil
}

// CleanExpiredRefreshTokens removes expired refresh tokens.
func (r *UserRepo) CleanExpiredRefreshTokens(ctx context.Context) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM refresh_tokens WHERE expires_at < NOW()`,
	)
	if err != nil {
		return 0, fmt.Errorf("clean expired refresh tokens: %w", err)
	}
	return tag.RowsAffected(), nil
}
