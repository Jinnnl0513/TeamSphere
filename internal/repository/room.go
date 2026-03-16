package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/teamsphere/server/internal/model"
)

type RoomRepo struct {
	db DBTX
}

func NewRoomRepo(db DBTX) *RoomRepo {
	return &RoomRepo{db: db}
}

// 閳光偓閳光偓閳光偓 Room CRUD 閳光偓閳光偓閳光偓

// Create inserts a new room and adds the creator as owner member. Returns the room.
func (r *RoomRepo) Create(ctx context.Context, name, description string, creatorID int64) (*model.Room, error) {
	room := &model.Room{}
	if err := withTx(ctx, r.db, func(db DBTX) error {
		if err := db.QueryRow(ctx,
			`INSERT INTO rooms (name, description, creator_id)
			 VALUES ($1, $2, $3)
			 RETURNING id, name, description, creator_id, created_at, updated_at`,
			name, description, creatorID,
		).Scan(&room.ID, &room.Name, &room.Description, &room.CreatorID, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return fmt.Errorf("insert room: %w", err)
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, 'owner')`,
			room.ID, creatorID,
		); err != nil {
			return fmt.Errorf("insert owner member: %w", err)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return room, nil
}

// GetByID retrieves a room by ID.
func (r *RoomRepo) GetByID(ctx context.Context, id int64) (*model.Room, error) {
	room := &model.Room{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, description, creator_id, created_at, updated_at
		 FROM rooms WHERE id = $1`, id,
	).Scan(&room.ID, &room.Name, &room.Description, &room.CreatorID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get room: %w", err)
	}
	return room, nil
}

// Update modifies a room's name and description.
func (r *RoomRepo) Update(ctx context.Context, id int64, name, description string) (*model.Room, error) {
	room := &model.Room{}
	err := r.db.QueryRow(ctx,
		`UPDATE rooms SET name = $1, description = $2, updated_at = NOW() WHERE id = $3
		 RETURNING id, name, description, creator_id, created_at, updated_at`,
		name, description, id,
	).Scan(&room.ID, &room.Name, &room.Description, &room.CreatorID, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update room: %w", err)
	}
	return room, nil
}

// Delete removes a room and all related data (CASCADE).
func (r *RoomRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM rooms WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete room: %w", err)
	}
	return nil
}

// ListByUser returns all rooms that a user has joined.
func (r *RoomRepo) ListByUser(ctx context.Context, userID int64) ([]*model.Room, error) {
	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.name, r.description, r.creator_id, r.created_at, r.updated_at
		 FROM rooms r
		 JOIN room_members rm ON rm.room_id = r.id
		 WHERE rm.user_id = $1
		 ORDER BY r.id`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list rooms by user: %w", err)
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		room := &model.Room{}
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.CreatorID, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// DiscoverAll returns latest discoverable public rooms.
func (r *RoomRepo) DiscoverAll(ctx context.Context) ([]*model.Room, error) {
	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.name, r.description, r.creator_id, r.created_at, r.updated_at
		 FROM rooms r
		 JOIN room_settings s ON s.room_id = r.id
		 WHERE s.is_public = TRUE
		 ORDER BY r.created_at DESC LIMIT 50`,
	)
	if err != nil {
		return nil, fmt.Errorf("discover rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*model.Room
	for rows.Next() {
		room := &model.Room{}
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.CreatorID, &room.CreatedAt, &room.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, room)
	}
	return rooms, nil
}

// 閳光偓閳光偓閳光偓 Room Members 閳光偓閳光偓閳光偓

// GetMember retrieves a single room member record.
func (r *RoomRepo) GetMember(ctx context.Context, roomID, userID int64) (*model.RoomMember, error) {
	m := &model.RoomMember{}
	err := r.db.QueryRow(ctx,
		`SELECT room_id, user_id, role, muted_until, joined_at
		 FROM room_members WHERE room_id = $1 AND user_id = $2`,
		roomID, userID,
	).Scan(&m.RoomID, &m.UserID, &m.Role, &m.MutedUntil, &m.JoinedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get member: %w", err)
	}
	return m, nil
}

// IsMember checks if a user is currently a member of the room.
func (r *RoomRepo) IsMember(ctx context.Context, roomID, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2)`,
		roomID, userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check member exists: %w", err)
	}
	return exists, nil
}

// HasOwnedRooms reports whether the user is currently the owner of any room.
func (r *RoomRepo) HasOwnedRooms(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM room_members
			WHERE user_id = $1 AND role = 'owner'
		)`,
		userID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check owned rooms: %w", err)
	}
	return exists, nil
}

// MemberInfo is a room member with user details attached.
type MemberInfo struct {
	model.RoomMember
	User model.UserInfo `json:"user"`
}

// ListMembers returns all members of a room with their user info.
func (r *RoomRepo) ListMembers(ctx context.Context, roomID int64) ([]*MemberInfo, error) {
	rows, err := r.db.Query(ctx,
		`SELECT rm.room_id, rm.user_id, rm.role, rm.muted_until, rm.joined_at,
		        u.id, u.username, u.avatar_url
		 FROM room_members rm
		 JOIN users u ON u.id = rm.user_id
		 WHERE rm.room_id = $1 AND u.deleted_at IS NULL
		 ORDER BY rm.joined_at`, roomID,
	)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []*MemberInfo
	for rows.Next() {
		mi := &MemberInfo{}
		if err := rows.Scan(
			&mi.RoomID, &mi.UserID, &mi.Role, &mi.MutedUntil, &mi.JoinedAt,
			&mi.User.ID, &mi.User.Username, &mi.User.AvatarURL,
		); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		members = append(members, mi)
	}
	return members, nil
}

// ListMemberIDs returns all member user IDs for a room.
func (r *RoomRepo) ListMemberIDs(ctx context.Context, roomID int64) ([]int64, error) {
	rows, err := r.db.Query(ctx, `SELECT user_id FROM room_members WHERE room_id = $1`, roomID)
	if err != nil {
		return nil, fmt.Errorf("list member ids: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan member id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate member ids: %w", err)
	}
	return ids, nil
}

// AddMember inserts a room member.
func (r *RoomRepo) AddMember(ctx context.Context, roomID, userID int64, role string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO room_members (room_id, user_id, role) VALUES ($1, $2, $3)`,
		roomID, userID, role,
	)
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

// RemoveMember removes a member from a room.
func (r *RoomRepo) RemoveMember(ctx context.Context, roomID, userID int64) error {
	_, err := r.db.Exec(ctx,
		`DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`,
		roomID, userID,
	)
	if err != nil {
		return fmt.Errorf("remove member: %w", err)
	}
	return nil
}

// UpdateMemberRole changes a member's room role.
func (r *RoomRepo) UpdateMemberRole(ctx context.Context, roomID, userID int64, role string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room_members SET role = $1 WHERE room_id = $2 AND user_id = $3`,
		role, roomID, userID,
	)
	if err != nil {
		return fmt.Errorf("update member role: %w", err)
	}
	return nil
}

// SetMuted sets or clears the muted_until timestamp for a member.
// Pass nil to unmute.
func (r *RoomRepo) SetMuted(ctx context.Context, roomID, userID int64, mutedUntil *time.Time) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room_members SET muted_until = $1 WHERE room_id = $2 AND user_id = $3`,
		mutedUntil, roomID, userID,
	)
	if err != nil {
		return fmt.Errorf("set muted: %w", err)
	}
	return nil
}

// TransferOwner atomically changes ownership: old owner 閳?admin, new owner 閳?owner.
func (r *RoomRepo) TransferOwner(ctx context.Context, roomID, oldOwnerID, newOwnerID int64) error {
	return withTx(ctx, r.db, func(db DBTX) error {
		if _, err := db.Exec(ctx,
			`UPDATE room_members SET role = 'admin' WHERE room_id = $1 AND user_id = $2`,
			roomID, oldOwnerID,
		); err != nil {
			return fmt.Errorf("demote old owner: %w", err)
		}
		if _, err := db.Exec(ctx,
			`UPDATE room_members SET role = 'owner' WHERE room_id = $1 AND user_id = $2`,
			roomID, newOwnerID,
		); err != nil {
			return fmt.Errorf("promote new owner: %w", err)
		}
		return nil
	})
}

// 閳光偓閳光偓閳光偓 Room Invites 閳光偓閳光偓閳光偓

// InviteInfo is an invite with room and inviter details.
type InviteInfo struct {
	model.RoomInvite
	RoomName        string `json:"room_name"`
	InviterUsername string `json:"inviter_username"`
}

// CreateInvite inserts a new room invite.
func (r *RoomRepo) CreateInvite(ctx context.Context, roomID, inviterID, inviteeID int64) (*model.RoomInvite, error) {
	inv := &model.RoomInvite{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO room_invites (room_id, inviter_id, invitee_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, room_id, inviter_id, invitee_id, status, created_at`,
		roomID, inviterID, inviteeID,
	).Scan(&inv.ID, &inv.RoomID, &inv.InviterID, &inv.InviteeID, &inv.Status, &inv.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create invite: %w", err)
	}
	return inv, nil
}

// GetInviteByID retrieves an invite by ID.
func (r *RoomRepo) GetInviteByID(ctx context.Context, id int64) (*model.RoomInvite, error) {
	inv := &model.RoomInvite{}
	err := r.db.QueryRow(ctx,
		`SELECT id, room_id, inviter_id, invitee_id, status, created_at
		 FROM room_invites WHERE id = $1`, id,
	).Scan(&inv.ID, &inv.RoomID, &inv.InviterID, &inv.InviteeID, &inv.Status, &inv.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get invite: %w", err)
	}
	return inv, nil
}

// HasPendingInvite checks if a pending invite already exists for this room+invitee.
func (r *RoomRepo) HasPendingInvite(ctx context.Context, roomID, inviteeID int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM room_invites WHERE room_id = $1 AND invitee_id = $2 AND status = 'pending')`,
		roomID, inviteeID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check pending invite: %w", err)
	}
	return exists, nil
}

// UpdateInviteStatus updates an invite's status (accepted/declined).
func (r *RoomRepo) UpdateInviteStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE room_invites SET status = $1 WHERE id = $2`,
		status, id,
	)
	if err != nil {
		return fmt.Errorf("update invite status: %w", err)
	}
	return nil
}

// ListPendingInvitesByUser returns all pending invites for a user, with room and inviter info.
func (r *RoomRepo) ListPendingInvitesByUser(ctx context.Context, userID int64) ([]*InviteInfo, error) {
	rows, err := r.db.Query(ctx,
		`SELECT ri.id, ri.room_id, ri.inviter_id, ri.invitee_id, ri.status, ri.created_at,
		        rm.name, u.username
		 FROM room_invites ri
		 JOIN rooms rm ON rm.id = ri.room_id
		 JOIN users u ON u.id = ri.inviter_id
		 WHERE ri.invitee_id = $1 AND ri.status = 'pending'
		 ORDER BY ri.created_at DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list pending invites: %w", err)
	}
	defer rows.Close()

	var invites []*InviteInfo
	for rows.Next() {
		ii := &InviteInfo{}
		if err := rows.Scan(
			&ii.ID, &ii.RoomID, &ii.InviterID, &ii.InviteeID, &ii.Status, &ii.CreatedAt,
			&ii.RoomName, &ii.InviterUsername,
		); err != nil {
			return nil, fmt.Errorf("scan invite: %w", err)
		}
		invites = append(invites, ii)
	}
	return invites, nil
}

// 閳光偓閳光偓閳光偓 Friendship check (needed for invite validation) 閳光偓閳光偓閳光偓

// AreFriends checks if two users are friends (accepted friendship in either direction).
func (r *RoomRepo) AreFriends(ctx context.Context, userA, userB int64) (bool, error) {
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

// CountMembers returns the number of members in a room.
func (r *RoomRepo) CountMembers(ctx context.Context, roomID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM room_members WHERE room_id = $1`, roomID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count members: %w", err)
	}
	return count, nil
}

// ListMemberUsernames returns a map of username閳�鎶瞫erID for all members of a room.
func (r *RoomRepo) ListMemberUsernames(ctx context.Context, roomID int64) (map[string]int64, error) {
	rows, err := r.db.Query(ctx,
		`SELECT u.username, u.id
		 FROM room_members rm
		 JOIN users u ON u.id = rm.user_id
		 WHERE rm.room_id = $1`, roomID,
	)
	if err != nil {
		return nil, fmt.Errorf("list member usernames: %w", err)
	}
	defer rows.Close()

	m := make(map[string]int64)
	for rows.Next() {
		var username string
		var userID int64
		if err := rows.Scan(&username, &userID); err != nil {
			return nil, fmt.Errorf("scan member username: %w", err)
		}
		m[username] = userID
	}
	return m, nil
}

// 閳光偓閳光偓閳光偓 Admin queries 閳光偓閳光偓閳光偓

// CountAll returns the total number of rooms.
func (r *RoomRepo) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM rooms`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count all rooms: %w", err)
	}
	return count, nil
}

// RoomWithMemberCount is a room with its member count.
type RoomWithMemberCount struct {
	model.Room
	MemberCount int `json:"member_count"`
}

// ListAllRooms returns paginated rooms with member counts.
func (r *RoomRepo) ListAllRooms(ctx context.Context, offset, limit int) ([]*RoomWithMemberCount, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	rows, err := r.db.Query(ctx,
		`SELECT r.id, r.name, r.description, r.creator_id, r.created_at, r.updated_at,
		        (SELECT COUNT(*) FROM room_members rm WHERE rm.room_id = r.id) AS member_count
		 FROM rooms r
		 ORDER BY r.id ASC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list all rooms: %w", err)
	}
	defer rows.Close()

	var rooms []*RoomWithMemberCount
	for rows.Next() {
		rc := &RoomWithMemberCount{}
		if err := rows.Scan(
			&rc.ID, &rc.Name, &rc.Description, &rc.CreatorID, &rc.CreatedAt, &rc.UpdatedAt,
			&rc.MemberCount,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rooms = append(rooms, rc)
	}
	return rooms, rows.Err()
}
