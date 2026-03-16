package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
)

type FriendshipRepo struct {
	db DBTX
}

func NewFriendshipRepo(db DBTX) *FriendshipRepo {
	return &FriendshipRepo{db: db}
}

// Create inserts a friendship request (status=pending).
func (r *FriendshipRepo) Create(ctx context.Context, userID, friendID int64) (*model.Friendship, error) {
	f := &model.Friendship{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO friendships (user_id, friend_id)
		 VALUES ($1, $2)
		 RETURNING id, user_id, friend_id, status, created_at, updated_at`,
		userID, friendID,
	).Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create friendship: %w", err)
	}
	return f, nil
}

// GetByID retrieves a friendship by ID.
func (r *FriendshipRepo) GetByID(ctx context.Context, id int64) (*model.Friendship, error) {
	f := &model.Friendship{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, friend_id, status, created_at, updated_at
		 FROM friendships WHERE id = $1`, id,
	).Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get friendship: %w", err)
	}
	return f, nil
}

// CheckExisting checks if a friendship (in either direction) already exists between two users.
// Returns the friendship if found, nil otherwise.
func (r *FriendshipRepo) CheckExisting(ctx context.Context, userA, userB int64) (*model.Friendship, error) {
	f := &model.Friendship{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, friend_id, status, created_at, updated_at
		 FROM friendships
		 WHERE (user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1)`,
		userA, userB,
	).Scan(&f.ID, &f.UserID, &f.FriendID, &f.Status, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("check existing friendship: %w", err)
	}
	return f, nil
}

// Accept updates a friendship status to accepted.
func (r *FriendshipRepo) Accept(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE friendships SET status = 'accepted', updated_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("accept friendship: %w", err)
	}
	return nil
}

// Delete removes a friendship record.
func (r *FriendshipRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM friendships WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete friendship: %w", err)
	}
	return nil
}

// FriendInfo is a friend with user details attached.
type FriendInfo struct {
	FriendshipID int64          `json:"friendship_id"`
	User         model.UserInfo `json:"user"`
}

// ListFriends returns all accepted friends for a user.
func (r *FriendshipRepo) ListFriends(ctx context.Context, userID int64) ([]*FriendInfo, error) {
	rows, err := r.db.Query(ctx,
		`SELECT f.id, u.id, u.username, u.avatar_url
		 FROM friendships f
		 JOIN users u ON u.id = CASE WHEN f.user_id = $1 THEN f.friend_id ELSE f.user_id END
		 WHERE (f.user_id = $1 OR f.friend_id = $1) AND f.status = 'accepted'
		 ORDER BY u.username`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}
	defer rows.Close()

	var friends []*FriendInfo
	for rows.Next() {
		fi := &FriendInfo{}
		if err := rows.Scan(&fi.FriendshipID, &fi.User.ID, &fi.User.Username, &fi.User.AvatarURL); err != nil {
			return nil, fmt.Errorf("scan friend: %w", err)
		}
		friends = append(friends, fi)
	}
	return friends, nil
}

// FriendRequestInfo is a pending request with requester details.
type FriendRequestInfo struct {
	model.Friendship
	From model.UserInfo `json:"from"`
}

// ListPendingRequests returns all pending friend requests received by a user.
func (r *FriendshipRepo) ListPendingRequests(ctx context.Context, userID int64) ([]*FriendRequestInfo, error) {
	rows, err := r.db.Query(ctx,
		`SELECT f.id, f.user_id, f.friend_id, f.status, f.created_at, f.updated_at,
		        u.id, u.username, u.avatar_url
		 FROM friendships f
		 JOIN users u ON u.id = f.user_id
		 WHERE f.friend_id = $1 AND f.status = 'pending'
		 ORDER BY f.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pending requests: %w", err)
	}
	defer rows.Close()

	var requests []*FriendRequestInfo
	for rows.Next() {
		ri := &FriendRequestInfo{}
		if err := rows.Scan(
			&ri.ID, &ri.UserID, &ri.FriendID, &ri.Status, &ri.CreatedAt, &ri.UpdatedAt,
			&ri.From.ID, &ri.From.Username, &ri.From.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf("scan request: %w", err)
		}
		requests = append(requests, ri)
	}
	return requests, nil
}

// AreFriends checks if two users are accepted friends.
func (r *FriendshipRepo) AreFriends(ctx context.Context, userA, userB int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM friendships
			WHERE status = 'accepted'
			  AND ((user_id = $1 AND friend_id = $2) OR (user_id = $2 AND friend_id = $1))
		)`, userA, userB,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check friendship: %w", err)
	}
	return exists, nil
}

// ListFriendIDs returns all accepted friend user IDs for a user.
func (r *FriendshipRepo) ListFriendIDs(ctx context.Context, userID int64) ([]int64, error) {
	rows, err := r.db.Query(ctx,
		`SELECT CASE WHEN user_id = $1 THEN friend_id ELSE user_id END
		 FROM friendships
		 WHERE (user_id = $1 OR friend_id = $1) AND status = 'accepted'`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list friend ids: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan friend id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// SearchUsers searches for users by username prefix (for adding friends).
func (r *FriendshipRepo) SearchUsers(ctx context.Context, query string, excludeUserID int64, limit int) ([]*model.UserInfo, error) {
	if limit <= 0 {
		limit = 20
	}
	// Escape SQL LIKE special characters to prevent wildcard injection
	escapedQuery := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(query)
	rows, err := r.db.Query(ctx,
		`SELECT id, username, avatar_url
		 FROM users
		 WHERE username ILIKE $1 AND id != $2 AND deleted_at IS NULL
		 ORDER BY username
		 LIMIT $3`,
		escapedQuery+"%", excludeUserID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	var users []*model.UserInfo
	for rows.Next() {
		u := &model.UserInfo{}
		if err := rows.Scan(&u.ID, &u.Username, &u.AvatarURL); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, nil
}
