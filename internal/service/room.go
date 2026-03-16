package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/teamsphere/server/internal/contract/tx"
	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/pkg/authutil"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
	"github.com/jackc/pgx/v5/pgconn"
)

// roleLevel returns a numeric level for room role comparison.
// Higher number = higher privilege. Delegates to authutil.RoleLevel.
func roleLevel(role string) int { return authutil.RoleLevel(role) }

// systemRoleLevel returns privilege level for system roles. Delegates to authutil.SystemRoleLevel.
func systemRoleLevel(role string) int { return authutil.SystemRoleLevel(role) }

type RoomService struct {
	roomRepo repository.RoomRepository
	userRepo repository.UserRepository
	settingsRepo repository.RoomSettingsRepository
	hub      *ws.Hub
	txMgr    tx.Manager
}

func NewRoomService(roomRepo repository.RoomRepository, userRepo repository.UserRepository, settingsRepo repository.RoomSettingsRepository, hub *ws.Hub, txMgr tx.Manager) *RoomService {
	return &RoomService{roomRepo: roomRepo, userRepo: userRepo, settingsRepo: settingsRepo, hub: hub, txMgr: txMgr}
}

// Create creates a new room. Any user can create.
func (s *RoomService) Create(ctx context.Context, userID int64, name, description string) (*model.Room, error) {
	room, err := s.roomRepo.Create(ctx, name, description, userID)
	if err != nil {
		// Check for unique violation on name
		if isDuplicateKey(err) {
			return nil, ErrRoomNameTaken
		}
		return nil, err
	}
	if s.settingsRepo != nil {
		_ = s.settingsRepo.CreateDefault(ctx, room.ID)
	}
	return room, nil
}

// GetByID returns room details. Caller must be a member.
func (s *RoomService) GetByID(ctx context.Context, userID, roomID int64) (*model.Room, error) {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotRoomMember
	}
	return room, nil
}

// Update modifies room name/description. Requires room admin+ or system admin+.
func (s *RoomService) Update(ctx context.Context, userID int64, systemRole string, roomID int64, name, description string) (*model.Room, error) {
	if err := s.requireRoomRole(ctx, userID, systemRole, roomID, "admin"); err != nil {
		return nil, err
	}
	room, err := s.roomRepo.Update(ctx, roomID, name, description)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, ErrRoomNameTaken
		}
		return nil, err
	}

	// WS: notify room members
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomUpdated,
		RoomID: roomID,
		Data: map[string]any{
			"room_id":     roomID,
			"name":        room.Name,
			"description": room.Description,
		},
	})

	return room, nil
}

// Delete removes a room. Requires room owner or system admin+.
func (s *RoomService) Delete(ctx context.Context, userID int64, systemRole string, roomID int64) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	// System admin+ can always delete
	if systemRoleLevel(systemRole) >= 2 {
		// Notify before deleting (room data needed for broadcast)
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeRoomDeleted,
			RoomID: roomID,
			Data:   map[string]any{"room_id": roomID},
		})
		return s.roomRepo.Delete(ctx, roomID)
	}

	// Room owner can delete
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil || member.Role != "owner" {
		return ErrNoPermission
	}

	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomDeleted,
		RoomID: roomID,
		Data:   map[string]any{"room_id": roomID},
	})
	return s.roomRepo.Delete(ctx, roomID)
}

// ListByUser returns all rooms a user has joined.
func (s *RoomService) ListByUser(ctx context.Context, userID int64) ([]*model.Room, error) {
	return s.roomRepo.ListByUser(ctx, userID)
}

// DiscoverAll returns all public rooms for exploration.
func (s *RoomService) DiscoverAll(ctx context.Context) ([]*model.Room, error) {
	return s.roomRepo.DiscoverAll(ctx)
}

// JoinRoom allows a user to voluntarily join a public room.
func (s *RoomService) JoinRoom(ctx context.Context, userID, roomID int64) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	isMem, err := s.roomRepo.IsMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if isMem {
		return ErrAlreadyMember
	}
	if s.settingsRepo != nil {
		settings, err := s.settingsRepo.GetByRoomID(ctx, roomID)
		if err != nil {
			return err
		}
		if settings != nil {
			if !settings.IsPublic {
				return ErrJoinInviteOnly
			}
			if settings.RequireApproval {
				if _, err := s.settingsRepo.CreateJoinRequest(ctx, roomID, userID, nil); err != nil {
					return ErrJoinRequestPending
				}
				return ErrJoinApprovalRequired
			}
		}
	}

	if err := s.roomRepo.AddMember(ctx, roomID, userID, "member"); err != nil {
		return err
	}

	user, err2 := s.userRepo.GetByID(ctx, userID)
	if err2 != nil {
		slog.Error("failed to get user for join notification", "user_id", userID, "error", err2)
	}
	if user != nil {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeRoomMemberJoined,
			RoomID: roomID,
			Data: map[string]any{
				"room_id": roomID,
				"user":    user.ToInfo(),
			},
		})
	}
	// #6: Keep roomMembers cache in sync
	s.hub.NotifyMemberJoined(roomID, userID)

	return nil
}

// ListMembers returns all members of a room. Caller must be a member.
func (s *RoomService) ListMembers(ctx context.Context, userID, roomID int64) ([]*repository.MemberInfo, error) {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotRoomMember
	}
	return s.roomRepo.ListMembers(ctx, roomID)
}

// InviteFriend invites a friend to the room. Requires room admin+.
func (s *RoomService) InviteFriend(ctx context.Context, userID int64, systemRole string, roomID, friendUserID int64) (*model.RoomInvite, error) {
	if err := s.requireRoomRole(ctx, userID, systemRole, roomID, "admin"); err != nil {
		return nil, err
	}

	// Check friendship
	friends, err := s.roomRepo.AreFriends(ctx, userID, friendUserID)
	if err != nil {
		return nil, err
	}
	if !friends {
		return nil, ErrInviteNotFriend
	}

	// Check if already a member
	existingMember, err := s.roomRepo.GetMember(ctx, roomID, friendUserID)
	if err != nil {
		return nil, err
	}
	if existingMember != nil {
		return nil, ErrAlreadyMember
	}

	// Check if pending invite exists
	hasPending, err := s.roomRepo.HasPendingInvite(ctx, roomID, friendUserID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, ErrInvitePending
	}

	invite, err := s.roomRepo.CreateInvite(ctx, roomID, userID, friendUserID)
	if err != nil {
		return nil, err
	}

	// WS: notify invitee
	room, err2 := s.roomRepo.GetByID(ctx, roomID)
	if err2 != nil {
		slog.Error("failed to get room for invite notification", "room_id", roomID, "error", err2)
	}
	inviter, err3 := s.userRepo.GetByID(ctx, userID)
	if err3 != nil {
		slog.Error("failed to get inviter for invite notification", "user_id", userID, "error", err3)
	}
	if room != nil && inviter != nil {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeRoomInvite,
			UserID: friendUserID,
			Data: map[string]any{
				"invite_id": invite.ID,
				"room":      map[string]any{"id": room.ID, "name": room.Name},
				"inviter":   inviter.ToInfo(),
			},
		})
	}

	return invite, nil
}

// ListPendingInvites returns all pending invites for a user.
func (s *RoomService) ListPendingInvites(ctx context.Context, userID int64) ([]*repository.InviteInfo, error) {
	return s.roomRepo.ListPendingInvitesByUser(ctx, userID)
}

// RespondInvite accepts or declines an invite. Only the invitee can respond.
func (s *RoomService) RespondInvite(ctx context.Context, userID, inviteID int64, accept bool) error {
	var invite *model.RoomInvite
	if s.txMgr != nil {
		if err := s.txMgr.WithTx(ctx, func(t tx.Tx) error {
			var err error
			invite, err = t.RoomRepo().GetInviteByID(ctx, inviteID)
			if err != nil {
				return err
			}
			if invite == nil || invite.Status != "pending" {
				return ErrInviteNotFound
			}
			if invite.InviteeID != userID {
				return ErrNoPermission
			}
			if accept {
				if err := t.RoomRepo().UpdateInviteStatus(ctx, inviteID, "accepted"); err != nil {
					return err
				}
				if s.settingsRepo != nil {
					if settings, err := s.settingsRepo.GetByRoomID(ctx, invite.RoomID); err == nil && settings != nil && settings.RequireApproval {
						if _, err := s.settingsRepo.CreateJoinRequest(ctx, invite.RoomID, userID, nil); err != nil {
							return err
						}
						return nil
					}
				}
				if err := t.RoomRepo().AddMember(ctx, invite.RoomID, userID, "member"); err != nil {
					return err
				}
				return nil
			}
			return t.RoomRepo().UpdateInviteStatus(ctx, inviteID, "declined")
		}); err != nil {
			return err
		}
	} else {
		var err error
		invite, err = s.roomRepo.GetInviteByID(ctx, inviteID)
		if err != nil {
			return err
		}
		if invite == nil || invite.Status != "pending" {
			return ErrInviteNotFound
		}
		if invite.InviteeID != userID {
			return ErrNoPermission
		}
		if accept {
			if err := s.roomRepo.UpdateInviteStatus(ctx, inviteID, "accepted"); err != nil {
				return err
			}
			if s.settingsRepo != nil {
				if settings, err := s.settingsRepo.GetByRoomID(ctx, invite.RoomID); err == nil && settings != nil && settings.RequireApproval {
					if _, err := s.settingsRepo.CreateJoinRequest(ctx, invite.RoomID, userID, nil); err != nil {
						return err
					}
					goto doneInvite
				}
			}
			if err := s.roomRepo.AddMember(ctx, invite.RoomID, userID, "member"); err != nil {
				return err
			}
		} else {
			if err := s.roomRepo.UpdateInviteStatus(ctx, inviteID, "declined"); err != nil {
				return err
			}
		}
	}

doneInvite:
	if accept {
		accepter, err2 := s.userRepo.GetByID(ctx, userID)
		if err2 != nil {
			slog.Error("failed to get accepter for invite accept notification", "user_id", userID, "error", err2)
		}
		if accepter != nil {
			s.hub.SendAction(&ws.Action{
				Type:   ws.TypeRoomInviteAccepted,
				UserID: invite.InviterID,
				Data: map[string]any{
					"room_id": invite.RoomID,
					"user":    accepter.ToInfo(),
				},
			})
			s.hub.SendAction(&ws.Action{
				Type:   ws.TypeRoomMemberJoined,
				RoomID: invite.RoomID,
				Data: map[string]any{
					"room_id": invite.RoomID,
					"user":    accepter.ToInfo(),
				},
			})
			s.hub.NotifyMemberJoined(invite.RoomID, userID)
		}
		return nil
	}

	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomInviteDeclined,
		UserID: invite.InviterID,
		Data: map[string]any{
			"invite_id":  inviteID,
			"invitee_id": userID,
		},
	})
	return nil
}

// UpdateMemberRole changes a member's role. Only room owner can do this.
func (s *RoomService) UpdateMemberRole(ctx context.Context, userID int64, systemRole string, roomID, targetUserID int64, newRole string) error {
	// Only room owner (or system admin+) can change roles
	if systemRoleLevel(systemRole) < 2 {
		member, err := s.roomRepo.GetMember(ctx, roomID, userID)
		if err != nil {
			return err
		}
		if member == nil || member.Role != "owner" {
			return ErrNoPermission
		}
	}

	// Validate new role (can only set admin or member)
	if newRole != "admin" && newRole != "member" {
		return fmt.Errorf("invalid role: %s", newRole)
	}

	target, err := s.roomRepo.GetMember(ctx, roomID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNotMember
	}
	if target.Role == "owner" {
		return ErrCannotActOnHigher
	}

	if err := s.roomRepo.UpdateMemberRole(ctx, roomID, targetUserID, newRole); err != nil {
		return err
	}

	// WS: notify room members
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomMemberRoleChanged,
		RoomID: roomID,
		Data: map[string]any{
			"room_id": roomID,
			"user_id": targetUserID,
			"role":    newRole,
		},
	})

	return nil
}

// KickMember removes a member from the room. Requires room admin+ or system admin+.
func (s *RoomService) KickMember(ctx context.Context, userID int64, systemRole string, roomID, targetUserID int64) error {
	actorRoomLevel, err := s.getEffectiveRoomLevel(ctx, userID, systemRole, roomID)
	if err != nil {
		return err
	}
	if actorRoomLevel < roleLevel("admin") {
		return ErrNoPermission
	}

	target, err := s.roomRepo.GetMember(ctx, roomID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNotMember
	}

	// Cannot kick same or higher role (unless system admin+)
	if systemRoleLevel(systemRole) < 2 && roleLevel(target.Role) >= actorRoomLevel {
		return ErrCannotActOnHigher
	}

	if err := s.roomRepo.RemoveMember(ctx, roomID, targetUserID); err != nil {
		return err
	}

	kickData := map[string]any{
		"room_id": roomID,
		"user_id": targetUserID,
	}
	// Broadcast to all room viewers (catches kicked user if they are viewing the room)
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomMemberKicked,
		RoomID: roomID,
		Data:   kickData,
	})
	// Also send directly to the kicked user in case they are not currently viewing the room
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomMemberKicked,
		UserID: targetUserID,
		Data:   kickData,
	})
	// #6: Keep roomMembers cache in sync
	s.hub.NotifyMemberLeft(roomID, targetUserID)

	return nil
}

// MuteMember mutes a member for a given duration. Requires room admin+ or system admin+.
func (s *RoomService) MuteMember(ctx context.Context, userID int64, systemRole string, roomID, targetUserID int64, durationMinutes int) error {
	actorRoomLevel, err := s.getEffectiveRoomLevel(ctx, userID, systemRole, roomID)
	if err != nil {
		return err
	}
	if actorRoomLevel < roleLevel("admin") {
		return ErrNoPermission
	}

	target, err := s.roomRepo.GetMember(ctx, roomID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNotMember
	}

	if systemRoleLevel(systemRole) < 2 && roleLevel(target.Role) >= actorRoomLevel {
		return ErrCannotActOnHigher
	}

	mutedUntil := time.Now().Add(time.Duration(durationMinutes) * time.Minute)
	if err := s.roomRepo.SetMuted(ctx, roomID, targetUserID, &mutedUntil); err != nil {
		return err
	}

	targetUser, err2 := s.userRepo.GetByID(ctx, targetUserID)
	if err2 != nil {
		slog.Error("failed to get target user for mute notification", "user_id", targetUserID, "error", err2)
	}
	var targetUserInfo any
	if targetUser != nil {
		targetUserInfo = targetUser.ToInfo()
	}

	// WS: notify room members
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomMemberMuted,
		RoomID: roomID,
		Data: map[string]any{
			"room_id":     roomID,
			"user":        targetUserInfo,
			"muted_until": mutedUntil,
		},
	})

	return nil
}

// UnmuteMember removes mute from a member. Requires room admin+ or system admin+.
func (s *RoomService) UnmuteMember(ctx context.Context, userID int64, systemRole string, roomID, targetUserID int64) error {
	actorRoomLevel, err := s.getEffectiveRoomLevel(ctx, userID, systemRole, roomID)
	if err != nil {
		return err
	}
	if actorRoomLevel < roleLevel("admin") {
		return ErrNoPermission
	}

	target, err := s.roomRepo.GetMember(ctx, roomID, targetUserID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNotMember
	}

	if err := s.roomRepo.SetMuted(ctx, roomID, targetUserID, nil); err != nil {
		return err
	}

	targetUser, err2 := s.userRepo.GetByID(ctx, targetUserID)
	if err2 != nil {
		slog.Error("failed to get target user for unmute notification", "user_id", targetUserID, "error", err2)
	}
	var targetUserInfo any
	if targetUser != nil {
		targetUserInfo = targetUser.ToInfo()
	}

	// WS: notify room members
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomMemberUnmuted,
		RoomID: roomID,
		Data: map[string]any{
			"room_id": roomID,
			"user":    targetUserInfo,
		},
	})

	return nil
}

// LeaveRoom allows a member to leave a room. Owner cannot leave.
func (s *RoomService) LeaveRoom(ctx context.Context, userID, roomID int64) error {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}
	if member.Role == "owner" {
		return ErrCannotLeaveOwner
	}
	if err := s.roomRepo.RemoveMember(ctx, roomID, userID); err != nil {
		return err
	}

	// WS: notify room members
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomMemberLeft,
		RoomID: roomID,
		Data: map[string]any{
			"room_id": roomID,
			"user_id": userID,
		},
	})
	// #6: Keep roomMembers cache in sync
	s.hub.NotifyMemberLeft(roomID, userID)

	return nil
}

// TransferOwner transfers room ownership. Only room owner can do this.
func (s *RoomService) TransferOwner(ctx context.Context, userID, roomID, newOwnerID int64) error {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil || member.Role != "owner" {
		return ErrNoPermission
	}

	target, err := s.roomRepo.GetMember(ctx, roomID, newOwnerID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrTargetNotMember
	}

	if err := s.roomRepo.TransferOwner(ctx, roomID, userID, newOwnerID); err != nil {
		return err
	}

	// WS: notify room members
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomOwnerTransferred,
		RoomID: roomID,
		Data: map[string]any{
			"room_id":      roomID,
			"old_owner_id": userID,
			"new_owner_id": newOwnerID,
		},
	})

	return nil
}

// ─── helpers ───

// requireRoomRole checks that the user has at least minRole in the room,
// or has system admin+ privilege.
func (s *RoomService) requireRoomRole(ctx context.Context, userID int64, systemRole string, roomID int64, minRole string) error {
	// System admin+ bypasses room role check
	if systemRoleLevel(systemRole) >= 2 {
		// Still verify room exists
		room, err := s.roomRepo.GetByID(ctx, roomID)
		if err != nil {
			return err
		}
		if room == nil {
			return ErrRoomNotFound
		}
		return nil
	}

	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}
	if roleLevel(member.Role) < roleLevel(minRole) {
		return ErrNoPermission
	}
	return nil
}

// getEffectiveRoomLevel returns the effective room privilege level for an actor.
// System admin+ gets at least admin-level access.
func (s *RoomService) getEffectiveRoomLevel(ctx context.Context, userID int64, systemRole string, roomID int64) (int, error) {
	sysLevel := systemRoleLevel(systemRole)

	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return 0, err
	}

	roomLvl := 0
	if member != nil {
		roomLvl = roleLevel(member.Role)
	}

	// System admin+ gets at least admin level
	if sysLevel >= 2 && roomLvl < roleLevel("admin") {
		return roleLevel("admin"), nil
	}

	if roomLvl == 0 && sysLevel < 2 {
		return 0, ErrNotRoomMember
	}

	return roomLvl, nil
}

// isDuplicateKey checks if an error is a PostgreSQL unique violation (error code 23505).
func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
