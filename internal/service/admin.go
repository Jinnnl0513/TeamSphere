package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
)

// AdminService provides system administration functionality.
// All methods require the caller to have system admin+ role (enforced in handler).
type AdminService struct {
	userRepo     repository.UserRepository
	roomRepo     repository.RoomRepository
	messageRepo  repository.MessageRepository
	settingsRepo repository.SettingsRepository
	emailService *EmailService
	hub          *ws.Hub
}

func NewAdminService(
	userRepo repository.UserRepository,
	roomRepo repository.RoomRepository,
	messageRepo repository.MessageRepository,
	settingsRepo repository.SettingsRepository,
	emailService *EmailService,
	hub *ws.Hub,
) *AdminService {
	return &AdminService{
		userRepo:     userRepo,
		roomRepo:     roomRepo,
		messageRepo:  messageRepo,
		settingsRepo: settingsRepo,
		emailService: emailService,
		hub:          hub,
	}
}

// 鈹€鈹€鈹€ User management 鈹€鈹€鈹€

// ListUsers returns a paginated list of all users.
func (s *AdminService) ListUsers(ctx context.Context, page, pageSize int) ([]*model.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	users, err := s.userRepo.ListAll(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateUserRole changes a user's system role.
// Rules:
//   - Cannot change your own role
//   - Only owner can set another user as admin
//   - There can only be one system owner
//   - Valid roles: user, admin
func (s *AdminService) UpdateUserRole(ctx context.Context, actorID, targetID int64, actorRole, newRole string) error {
	if actorID == targetID {
		return ErrAdminSelfRole
	}
	if newRole != "user" && newRole != "admin" {
		return ErrAdminInvalidRole
	}

	target, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	// Cannot change the owner's role
	if target.Role == "owner" {
		return ErrAdminDeleteOwner
	}
	ownsRooms, err := s.roomRepo.HasOwnedRooms(ctx, targetID)
	if err != nil {
		return err
	}
	if ownsRooms {
		return ErrOwnsRooms
	}
	// Only owner can promote to admin
	if newRole == "admin" && actorRole != "owner" {
		return ErrNoPermission
	}

	return s.userRepo.UpdateRole(ctx, targetID, newRole)
}

// DeleteUser forcefully deletes a user (hard delete).
// Rules:
//   - Cannot delete yourself
//   - Cannot delete the system owner
func (s *AdminService) DeleteUser(ctx context.Context, actorID, targetID int64) error {
	if actorID == targetID {
		return ErrAdminSelfDelete
	}

	target, err := s.userRepo.GetByID(ctx, targetID)
	if err != nil {
		return err
	}
	if target == nil {
		return ErrUserNotFound
	}
	if target.Role == "owner" {
		return ErrAdminDeleteOwner
	}
	ownsRooms, err := s.roomRepo.HasOwnedRooms(ctx, targetID)
	if err != nil {
		return err
	}
	if ownsRooms {
		return ErrOwnsRooms
	}

	// Force disconnect the user's WebSocket connections first
	if s.hub != nil {
		s.hub.SendAction(&ws.Action{
			Type: "_force_disconnect",
			Data: targetID,
		})
	}

	return s.userRepo.HardDelete(ctx, targetID)
}

// 鈹€鈹€鈹€ Room management 鈹€鈹€鈹€

// ListRooms returns a paginated list of all rooms with member counts.
func (s *AdminService) ListRooms(ctx context.Context, page, pageSize int) ([]*repository.RoomWithMemberCount, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	rooms, err := s.roomRepo.ListAllRooms(ctx, offset, pageSize)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.roomRepo.CountAll(ctx)
	if err != nil {
		return nil, 0, err
	}

	return rooms, total, nil
}

// DeleteRoom deletes a room by admin.
func (s *AdminService) DeleteRoom(ctx context.Context, roomID int64) error {
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return ErrRoomNotFound
	}

	// Broadcast room_deleted to viewers first
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeRoomDeleted,
		RoomID: roomID,
		Data: map[string]any{
			"room_id": roomID,
		},
	})

	return s.roomRepo.Delete(ctx, roomID)
}

// 鈹€鈹€鈹€ System stats 鈹€鈹€鈹€

// SystemStats holds aggregate system statistics.
type SystemStats struct {
	TotalUsers  int64 `json:"total_users"`
	ActiveUsers int64 `json:"active_users"`
	TotalRooms  int64 `json:"total_rooms"`
	TotalMsgs   int64 `json:"total_messages"`
	TotalDMs    int64 `json:"total_dms"`
	OnlineUsers int   `json:"online_users"`
}

// GetStats returns system statistics.
func (s *AdminService) GetStats(ctx context.Context) (*SystemStats, error) {
	totalUsers, err := s.userRepo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	activeUsers, err := s.userRepo.CountActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("count active users: %w", err)
	}

	totalRooms, err := s.roomRepo.CountAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("count rooms: %w", err)
	}

	totalMsgs, err := s.messageRepo.CountMessages(ctx)
	if err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}

	totalDMs, err := s.messageRepo.CountDMs(ctx)
	if err != nil {
		return nil, fmt.Errorf("count dms: %w", err)
	}

	onlineCount := s.hub.OnlineCount()

	return &SystemStats{
		TotalUsers:  totalUsers,
		ActiveUsers: activeUsers,
		TotalRooms:  totalRooms,
		TotalMsgs:   totalMsgs,
		TotalDMs:    totalDMs,
		OnlineUsers: onlineCount,
	}, nil
}

// 鈹€鈹€鈹€ System settings 鈹€鈹€鈹€

// GetSettings returns all system settings (masking sensitive values).
func (s *AdminService) GetSettings(ctx context.Context) (map[string]string, error) {
	settings, err := s.settingsRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	// Mask sensitive values
	for k := range settings {
		if k == "email.password" {
			settings[k] = "********"
			continue
		}
		if strings.Contains(k, "client_secret") {
			settings[k] = "********"
		}
	}
	return settings, nil
}

// allowedSettingKeys defines which system settings keys can be modified via the admin API.
// This prevents malicious admins from injecting unknown keys into system_settings.
var allowedSettingKeys = map[string]bool{
	"registration.email_required": true,
	"registration.allow_register": true,
	"announcement":                true,
	"oauth.enabled":               true,
	"oauth.frontend_base_url":     true,
	"oauth.github.enabled":        true,
	"oauth.github.client_id":      true,
	"oauth.github.client_secret":  true,
	"oauth.github.redirect_url":   true,
	"oauth.github.scopes":         true,
	"oauth.google.enabled":        true,
	"oauth.google.client_id":      true,
	"oauth.google.client_secret":  true,
	"oauth.google.redirect_url":   true,
	"oauth.google.scopes":         true,
	"oauth.google.hosted_domain":  true,
	"oauth.oidc.enabled":          true,
	"oauth.oidc.client_id":        true,
	"oauth.oidc.client_secret":    true,
	"oauth.oidc.redirect_url":     true,
	"oauth.oidc.scopes":           true,
	"oauth.oidc.issuer_url":       true,
	"security.2fa_policy":         true,
}

// UpdateSettings updates system settings, only allowing whitelisted keys.
func (s *AdminService) UpdateSettings(ctx context.Context, updates map[string]string) error {
	for k, v := range updates {
		if !allowedSettingKeys[k] {
			return fmt.Errorf("setting key %q is not allowed", k)
		}
		if err := s.settingsRepo.Set(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}

// 鈹€鈹€鈹€ Email settings 鈹€鈹€鈹€

// GetEmailSettings returns the current email configuration.
func (s *AdminService) GetEmailSettings(ctx context.Context) (*EmailSettings, error) {
	return s.emailService.GetSettings(ctx)
}

// UpdateEmailSettings updates email configuration.
func (s *AdminService) UpdateEmailSettings(ctx context.Context, settings *EmailSettings) error {
	return s.emailService.UpdateSettings(ctx, settings)
}

// 鈹€鈹€鈹€ Announcements 鈹€鈹€鈹€

// SetAnnouncement stores a system announcement in system_settings.
func (s *AdminService) SetAnnouncement(ctx context.Context, content string) error {
	return s.settingsRepo.Set(ctx, "announcement", content)
}

// GetAnnouncement reads the current system announcement.
func (s *AdminService) GetAnnouncement(ctx context.Context) (string, error) {
	return s.settingsRepo.Get(ctx, "announcement")
}
