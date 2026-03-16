package service

import (
	"context"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
)

type UserService struct {
	userRepo    repository.UserRepository
	roomRepo    repository.RoomRepository
	authService *AuthService
	hub         *ws.Hub
}

func NewUserService(userRepo repository.UserRepository, roomRepo repository.RoomRepository, authService *AuthService, hub *ws.Hub) *UserService {
	return &UserService{userRepo: userRepo, roomRepo: roomRepo, authService: authService, hub: hub}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) UpdateBioAndColor(ctx context.Context, id int64, bio, profileColor string) (*model.User, error) {
	return s.userRepo.UpdateBioAndColor(ctx, id, bio, profileColor)
}

func (s *UserService) ChangePassword(ctx context.Context, id int64, oldPassword, newPassword string, claims *AuthClaims) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := verifyPassword(user.Password, oldPassword); err != nil {
		return err
	}

	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.userRepo.UpdatePassword(ctx, id, hash); err != nil {
		return err
	}

	if claims != nil {
		_ = s.authService.Logout(ctx, claims, "")
	}
	return nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, id int64, avatarURL string) (*model.User, error) {
	return s.userRepo.UpdateAvatar(ctx, id, avatarURL)
}

func (s *UserService) DeleteAccount(ctx context.Context, id int64, password string, claims *AuthClaims) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	if err := verifyPassword(user.Password, password); err != nil {
		return err
	}

	ownsRooms, err := s.roomRepo.HasOwnedRooms(ctx, id)
	if err != nil {
		return err
	}
	if ownsRooms {
		return ErrOwnsRooms
	}

	if err := s.authService.RevokeAllRefreshTokens(ctx, id); err != nil {
		return err
	}

	if err := s.userRepo.SoftDelete(ctx, id); err != nil {
		return err
	}

	if claims != nil {
		_ = s.authService.Logout(ctx, claims, "")
	}

	if s.hub != nil {
		s.hub.SendAction(&ws.Action{
			Type: "_force_disconnect",
			Data: id,
		})
	}

	return nil
}
