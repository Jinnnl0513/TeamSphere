package repository

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

// InviteLinkRepo handles all invite_links database operations.
type InviteLinkRepo struct {
	db DBTX
}

var (
	ErrInviteLinkNotFound = errors.New("invite link not found")
	ErrInviteLinkExpired  = errors.New("invite link expired")
	ErrInviteLinkMaxUses  = errors.New("invite link max uses reached")
)

type UseInviteLinkResult struct {
	Link  *model.InviteLink
	IsNew bool
}

func NewInviteLinkRepo(db DBTX) *InviteLinkRepo {
	return &InviteLinkRepo{db: db}
}

// generateCode generates a random 8-byte URL-safe code (11 base64url chars, capped to 8).
func generateCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// base64url without padding -> 8 chars
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Create inserts a new invite link. maxUses=0 means unlimited; expiresAt=nil means never.
func (r *InviteLinkRepo) Create(ctx context.Context, roomID, creatorID int64, maxUses int, expiresAt *time.Time) (*model.InviteLink, error) {
	// Retry up to 5 times on code collision (extremely rare with 6-byte random)
	for attempt := 0; attempt < 5; attempt++ {
		code, err := generateCode()
		if err != nil {
			return nil, fmt.Errorf("generate invite code: %w", err)
		}

		link := &model.InviteLink{}
		err = r.db.QueryRow(ctx,
			`INSERT INTO invite_links (code, room_id, creator_id, max_uses, expires_at)
			 VALUES ($1, $2, $3, $4, $5)
			 RETURNING id, code, room_id, creator_id, max_uses, uses, expires_at, created_at`,
			code, roomID, creatorID, maxUses, expiresAt,
		).Scan(&link.ID, &link.Code, &link.RoomID, &link.CreatorID,
			&link.MaxUses, &link.Uses, &link.ExpiresAt, &link.CreatedAt)
		if err != nil {
			// Retry on unique violation (code collision)
			var pgErr interface{ SQLState() string }
			if errors.As(err, &pgErr) && pgErr.SQLState() == "23505" {
				continue
			}
			return nil, fmt.Errorf("create invite link: %w", err)
		}
		return link, nil
	}
	return nil, fmt.Errorf("failed to generate unique invite code after 5 attempts")
}

// GetByCode retrieves an invite link by its short code.
func (r *InviteLinkRepo) GetByCode(ctx context.Context, code string) (*model.InviteLink, error) {
	link := &model.InviteLink{}
	err := r.db.QueryRow(ctx,
		`SELECT id, code, room_id, creator_id, max_uses, uses, expires_at, created_at
		 FROM invite_links WHERE code = $1`, code,
	).Scan(&link.ID, &link.Code, &link.RoomID, &link.CreatorID,
		&link.MaxUses, &link.Uses, &link.ExpiresAt, &link.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get invite link by code: %w", err)
	}
	return link, nil
}

// Use atomically validates a link, adds membership if needed, and increments
// usage only when the join actually happens.
func (r *InviteLinkRepo) Use(ctx context.Context, code string, userID int64) (*UseInviteLinkResult, error) {
	link := &model.InviteLink{}
	var result *UseInviteLinkResult
	if err := withTx(ctx, r.db, func(db DBTX) error {
		if err := db.QueryRow(ctx,
			`SELECT id, code, room_id, creator_id, max_uses, uses, expires_at, created_at
			 FROM invite_links
			 WHERE code = $1
			 FOR UPDATE`,
			code,
		).Scan(&link.ID, &link.Code, &link.RoomID, &link.CreatorID,
			&link.MaxUses, &link.Uses, &link.ExpiresAt, &link.CreatedAt); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInviteLinkNotFound
			}
			return fmt.Errorf("lock invite link by code: %w", err)
		}

		now := time.Now()
		if link.ExpiresAt != nil && now.After(*link.ExpiresAt) {
			return ErrInviteLinkExpired
		}
		if link.MaxUses > 0 && link.Uses >= link.MaxUses {
			return ErrInviteLinkMaxUses
		}

		var isMember bool
		if err := db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`,
			link.RoomID, userID,
		).Scan(&isMember); err != nil {
			return fmt.Errorf("check invite link membership: %w", err)
		}
		if isMember {
			result = &UseInviteLinkResult{Link: link, IsNew: false}
			return nil
		}

		if _, err := db.Exec(ctx,
			`INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, 'member')`,
			link.RoomID, userID,
		); err != nil {
			return fmt.Errorf("add invite link member: %w", err)
		}

		if _, err := db.Exec(ctx,
			`UPDATE invite_links SET uses = uses + 1 WHERE id = $1`,
			link.ID,
		); err != nil {
			return fmt.Errorf("increment invite link uses: %w", err)
		}
		link.Uses++
		result = &UseInviteLinkResult{Link: link, IsNew: true}
		return nil
	}); err != nil {
		return nil, err
	}

	return result, nil
}

// IncrementUses atomically increments the use count by 1.
func (r *InviteLinkRepo) IncrementUses(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE invite_links SET uses = uses + 1 WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("increment invite link uses: %w", err)
	}
	return nil
}

// ListByRoom returns all invite links for a room, including creator info.
func (r *InviteLinkRepo) ListByRoom(ctx context.Context, roomID int64) ([]*model.InviteLink, error) {
	rows, err := r.db.Query(ctx,
		`SELECT il.id, il.code, il.room_id, il.creator_id,
		        COALESCE(u.username, ''),
		        il.max_uses, il.uses, il.expires_at, il.created_at
		 FROM invite_links il
		 LEFT JOIN users u ON u.id = il.creator_id
		 WHERE il.room_id = $1 ORDER BY il.created_at DESC`, roomID)
	if err != nil {
		return nil, fmt.Errorf("list invite links: %w", err)
	}
	defer rows.Close()

	var links []*model.InviteLink
	for rows.Next() {
		l := &model.InviteLink{}
		if err := rows.Scan(&l.ID, &l.Code, &l.RoomID, &l.CreatorID,
			&l.CreatorName,
			&l.MaxUses, &l.Uses, &l.ExpiresAt, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan invite link: %w", err)
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

// Delete removes an invite link by ID.
func (r *InviteLinkRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM invite_links WHERE id = $1`, id)
	return err
}
