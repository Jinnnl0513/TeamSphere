package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
)

type RoomSettingsService struct {
	settingsRepo repository.RoomSettingsRepository
	roomRepo     repository.RoomRepository
	userRepo     repository.UserRepository
	hub          *ws.Hub
}

func NewRoomSettingsService(settingsRepo repository.RoomSettingsRepository, roomRepo repository.RoomRepository, userRepo repository.UserRepository, hub *ws.Hub) *RoomSettingsService {
	return &RoomSettingsService{settingsRepo: settingsRepo, roomRepo: roomRepo, userRepo: userRepo, hub: hub}
}

func (s *RoomSettingsService) GetSettings(ctx context.Context, roomID int64) (*model.RoomSettings, error) {
	if s.settingsRepo == nil {
		return nil, fmt.Errorf("settings repo not configured")
	}
	settings, err := s.settingsRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if settings == nil {
		if err := s.settingsRepo.CreateDefault(ctx, roomID); err != nil {
			return nil, err
		}
		return s.settingsRepo.GetByRoomID(ctx, roomID)
	}
	return settings, nil
}

func (s *RoomSettingsService) UpdateSettings(ctx context.Context, userID int64, systemRole string, roomID int64, next *model.RoomSettings) (*model.RoomSettings, error) {
	if err := s.requireManageSettings(ctx, userID, systemRole, roomID); err != nil {
		return nil, err
	}
	if err := s.settingsRepo.CreateDefault(ctx, roomID); err != nil {
		return nil, err
	}
	normalized := normalizeRoomSettings(next)
	updated, err := s.settingsRepo.Update(ctx, roomID, normalized)
	if err != nil {
		return nil, err
	}

	if s.settingsRepo != nil {
		_ = s.settingsRepo.CreateAuditLog(ctx, roomID, &userID, "room_settings.update", nil, updated)
	}

	if s.hub != nil {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeRoomUpdated,
			RoomID: roomID,
			Data: map[string]any{
				"room_id": roomID,
				"settings": updated,
			},
		})
	}

	return updated, nil
}

func (s *RoomSettingsService) ListPermissions(ctx context.Context, roomID int64) ([]*model.RoomRolePermission, error) {
	return s.settingsRepo.ListPermissions(ctx, roomID)
}

func (s *RoomSettingsService) UpdatePermissions(ctx context.Context, userID int64, systemRole string, roomID int64, perms []*model.RoomRolePermission) error {
	if err := s.requireManagePermissions(ctx, userID, systemRole, roomID); err != nil {
		return err
	}
	for _, p := range perms {
		p.Role = strings.ToLower(strings.TrimSpace(p.Role))
	}
	if err := s.settingsRepo.UpsertPermissions(ctx, roomID, perms); err != nil {
		return err
	}
	if s.settingsRepo != nil {
		_ = s.settingsRepo.CreateAuditLog(ctx, roomID, &userID, "room_permissions.update", nil, perms)
	}
	return nil
}

func (s *RoomSettingsService) ListJoinRequests(ctx context.Context, userID int64, systemRole string, roomID int64) ([]*model.RoomJoinRequest, error) {
	if err := s.requireManageMembers(ctx, userID, systemRole, roomID); err != nil {
		return nil, err
	}
	return s.settingsRepo.ListJoinRequests(ctx, roomID)
}

func (s *RoomSettingsService) ApproveJoinRequest(ctx context.Context, reviewerID int64, systemRole string, reqID int64) error {
	req, err := s.settingsRepo.GetJoinRequest(ctx, reqID)
	if err != nil {
		return err
	}
	if req == nil || req.Status != "pending" {
		return ErrInviteNotFound
	}
	if err := s.requireManageMembers(ctx, reviewerID, systemRole, req.RoomID); err != nil {
		return err
	}
	if err := s.roomRepo.AddMember(ctx, req.RoomID, req.UserID, "member"); err != nil {
		return err
	}
	if err := s.settingsRepo.UpdateJoinRequestStatus(ctx, reqID, "approved", reviewerID); err != nil {
		return err
	}
	user, _ := s.userRepo.GetByID(ctx, req.UserID)
	if s.hub != nil && user != nil {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeRoomMemberJoined,
			RoomID: req.RoomID,
			Data: map[string]any{
				"room_id": req.RoomID,
				"user":    user.ToInfo(),
			},
		})
		s.hub.NotifyMemberJoined(req.RoomID, req.UserID)
	}
	return nil
}

func (s *RoomSettingsService) RejectJoinRequest(ctx context.Context, reviewerID int64, systemRole string, reqID int64) error {
	req, err := s.settingsRepo.GetJoinRequest(ctx, reqID)
	if err != nil {
		return err
	}
	if req == nil || req.Status != "pending" {
		return ErrInviteNotFound
	}
	if err := s.requireManageMembers(ctx, reviewerID, systemRole, req.RoomID); err != nil {
		return err
	}
	return s.settingsRepo.UpdateJoinRequestStatus(ctx, reqID, "rejected", reviewerID)
}

func (s *RoomSettingsService) GetStatsSummary(ctx context.Context, userID int64, systemRole string, roomID int64, days int) (*repository.RoomStatsSummary, error) {
	if err := s.requireManageMembers(ctx, userID, systemRole, roomID); err != nil {
		return nil, err
	}
	if days <= 0 {
		days = 7
	}
	since := time.Now().AddDate(0, 0, -days)
	return s.settingsRepo.GetStatsSummary(ctx, roomID, since)
}

func (s *RoomSettingsService) requireManageSettings(ctx context.Context, userID int64, systemRole string, roomID int64) error {
	if systemRoleLevel(systemRole) >= 2 {
		return nil
	}
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}
	if member.Role == "owner" || member.Role == "admin" {
		return nil
	}
	return ErrNoPermission
}

func (s *RoomSettingsService) requireManagePermissions(ctx context.Context, userID int64, systemRole string, roomID int64) error {
	if systemRoleLevel(systemRole) >= 2 {
		return nil
	}
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}
	if member.Role == "owner" {
		return nil
	}
	return ErrNoPermission
}

func (s *RoomSettingsService) requireManageMembers(ctx context.Context, userID int64, systemRole string, roomID int64) error {
	if systemRoleLevel(systemRole) >= 2 {
		return nil
	}
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}
	if member.Role == "owner" || member.Role == "admin" {
		return nil
	}
	return ErrNoPermission
}

func normalizeRoomSettings(s *model.RoomSettings) *model.RoomSettings {
	if s == nil {
		return &model.RoomSettings{}
	}
	if s.SlowModeSeconds < 0 {
		s.SlowModeSeconds = 0
	}
	if s.SlowModeSeconds > 300 {
		s.SlowModeSeconds = 300
	}
	if s.MessageRetention < 0 {
		s.MessageRetention = 0
	}
	if s.MaxFileSizeMB < 0 {
		s.MaxFileSizeMB = 0
	}
	if s.PinLimit <= 0 {
		s.PinLimit = 50
	}
	if s.NotifyMode == "" {
		s.NotifyMode = "all"
	}
	if s.ContentFilterMode == "" {
		s.ContentFilterMode = "off"
	}
	if s.AntiSpamRate < 0 {
		s.AntiSpamRate = 0
	}
	if s.AntiSpamWindowSec < 0 {
		s.AntiSpamWindowSec = 0
	}
	return s
}
