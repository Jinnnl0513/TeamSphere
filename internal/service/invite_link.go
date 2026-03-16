package service

import (
	"context"
	"errors"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
)

var (
	ErrInviteLinkNotFound = errors.New("invite link not found or invalid")
	ErrInviteLinkExpired  = errors.New("invite link has expired")
	ErrInviteLinkMaxUses  = errors.New("invite link has reached max uses")
	ErrInviteLinkForbid   = errors.New("no permission to manage invite link")
)

// InviteLinkResult is returned when a user successfully uses an invite link.
type InviteLinkResult struct {
	Room  *model.Room `json:"room"`
	IsNew bool        `json:"is_new"` // false when user was already a member
}

// InviteLinkService handles all invite link business logic.
type InviteLinkService struct {
	inviteRepo repository.InviteLinkRepository
	roomRepo   repository.RoomRepository
	userRepo   repository.UserRepository
	hub        *ws.Hub
}

func NewInviteLinkService(
	inviteRepo repository.InviteLinkRepository,
	roomRepo repository.RoomRepository,
	userRepo repository.UserRepository,
	hub *ws.Hub,
) *InviteLinkService {
	return &InviteLinkService{
		inviteRepo: inviteRepo,
		roomRepo:   roomRepo,
		userRepo:   userRepo,
		hub:        hub,
	}
}

// CreateLink creates a new invite link for a room.
// Requires the caller to be at least a member of the room.
// maxUses=0 means unlimited; expiresHours=0 means never expires.
func (s *InviteLinkService) CreateLink(ctx context.Context, userID, roomID int64, maxUses, expiresHours int) (*model.InviteLink, error) {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotRoomMember
	}

	var expiresAt *time.Time
	if expiresHours > 0 {
		t := time.Now().Add(time.Duration(expiresHours) * time.Hour)
		expiresAt = &t
	}

	return s.inviteRepo.Create(ctx, roomID, userID, maxUses, expiresAt)
}

// GetLinkInfo returns public info about an invite link (room name etc.).
// It does not join the user; use UseLink for that.
func (s *InviteLinkService) GetLinkInfo(ctx context.Context, code string) (*model.InviteLink, *model.Room, error) {
	link, err := s.inviteRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, nil, err
	}
	if link == nil {
		return nil, nil, ErrInviteLinkNotFound
	}
	if err := s.validateLink(link); err != nil {
		return nil, nil, err
	}
	room, err := s.roomRepo.GetByID(ctx, link.RoomID)
	if err != nil {
		return nil, nil, err
	}
	if room == nil {
		return nil, nil, ErrRoomNotFound
	}
	return link, room, nil
}

// UseLink validates the link and adds the user to the room.
func (s *InviteLinkService) UseLink(ctx context.Context, userID int64, code string) (*InviteLinkResult, error) {
	useResult, err := s.inviteRepo.Use(ctx, code, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInviteLinkNotFound):
			return nil, ErrInviteLinkNotFound
		case errors.Is(err, repository.ErrInviteLinkExpired):
			return nil, ErrInviteLinkExpired
		case errors.Is(err, repository.ErrInviteLinkMaxUses):
			return nil, ErrInviteLinkMaxUses
		default:
			return nil, err
		}
	}

	room, err := s.roomRepo.GetByID(ctx, useResult.Link.RoomID)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, ErrRoomNotFound
	}
	if !useResult.IsNew {
		return &InviteLinkResult{Room: room, IsNew: false}, nil
	}

	user, _ := s.userRepo.GetByID(ctx, userID)
	if user != nil {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeRoomMemberJoined,
			RoomID: useResult.Link.RoomID,
			Data: map[string]any{
				"room_id": useResult.Link.RoomID,
				"user":    user.ToInfo(),
			},
		})
	}
	// Keep roomMembers cache in sync.
	s.hub.NotifyMemberJoined(useResult.Link.RoomID, userID)

	return &InviteLinkResult{Room: room, IsNew: true}, nil
}

// ListLinks returns all invite links for a room. Caller must be a member.
func (s *InviteLinkService) ListLinks(ctx context.Context, userID, roomID int64) ([]*model.InviteLink, error) {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotRoomMember
	}
	return s.inviteRepo.ListByRoom(ctx, roomID)
}

// DeleteLink removes an invite link. Creator or room admin+ can delete.
func (s *InviteLinkService) DeleteLink(ctx context.Context, userID, roomID, linkID int64, systemRole string) error {
	links, err := s.inviteRepo.ListByRoom(ctx, roomID)
	if err != nil {
		return err
	}
	var target *model.InviteLink
	for _, l := range links {
		if l.ID == linkID {
			target = l
			break
		}
	}
	if target == nil {
		return ErrInviteLinkNotFound
	}

	// Creator can always delete their own link; room admin+ or system admin can delete others'.
	if target.CreatorID != userID {
		member, err := s.roomRepo.GetMember(ctx, roomID, userID)
		if err != nil {
			return err
		}
		isAdmin := systemRoleLevel(systemRole) >= 2
		isRoomAdmin := member != nil && roleLevel(member.Role) >= roleLevel("admin")
		if !isAdmin && !isRoomAdmin {
			return ErrInviteLinkForbid
		}
	}

	return s.inviteRepo.Delete(ctx, linkID)
}

// validateLink checks expiry and max-uses constraints.
func (s *InviteLinkService) validateLink(link *model.InviteLink) error {
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return ErrInviteLinkExpired
	}
	if link.MaxUses > 0 && link.Uses >= link.MaxUses {
		return ErrInviteLinkMaxUses
	}
	return nil
}
