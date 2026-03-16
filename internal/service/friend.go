package service

import (
	"context"
	"log/slog"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
)

type FriendService struct {
	friendRepo repository.FriendshipRepository
	userRepo   repository.UserRepository
	hub        *ws.Hub
	notificationService *NotificationService
}

func NewFriendService(friendRepo repository.FriendshipRepository, userRepo repository.UserRepository, hub *ws.Hub, notificationService *NotificationService) *FriendService {
	return &FriendService{friendRepo: friendRepo, userRepo: userRepo, hub: hub, notificationService: notificationService}
}

// SendRequest sends a friend request from userID to targetUserID.
func (s *FriendService) SendRequest(ctx context.Context, userID, targetUserID int64) (*model.Friendship, error) {
	if userID == targetUserID {
		return nil, ErrFriendSelf
	}

	// Check if friendship already exists in either direction
	existing, err := s.friendRepo.CheckExisting(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if existing.Status == "accepted" {
			return nil, ErrFriendExists
		}
		return nil, ErrFriendReqPending
	}

	friendship, err := s.friendRepo.Create(ctx, userID, targetUserID)
	if err != nil {
		return nil, err
	}

	// WS notification: notify target user of friend request
	sender, err2 := s.userRepo.GetByID(ctx, userID)
	if err2 != nil {
		slog.Error("failed to get sender for friend request notification", "user_id", userID, "error", err2)
	}
	if sender != nil {
		if s.notificationService != nil {
			_, _ = s.notificationService.Create(ctx, &model.Notification{
				UserID: targetUserID,
				Type:   "friend_request",
				Title:  "新的好友请求",
				Body:   sender.Username + " 向你发送了好友请求",
				RefID:  friendship.ID,
			})
		}
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeFriendRequest,
			UserID: targetUserID,
			Data: map[string]any{
				"request_id": friendship.ID,
				"from":       sender.ToInfo(),
			},
		})
	}

	return friendship, nil
}

// AcceptRequest accepts a pending friend request. Only the receiver (friend_id) can accept.
func (s *FriendService) AcceptRequest(ctx context.Context, userID, requestID int64) error {
	friendship, err := s.friendRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if friendship == nil || friendship.Status != "pending" {
		return ErrFriendReqNotFound
	}
	if friendship.FriendID != userID {
		return ErrNoPermission
	}

	if err := s.friendRepo.Accept(ctx, requestID); err != nil {
		return err
	}

	// WS notification: notify requester that their request was accepted
	accepter, err2 := s.userRepo.GetByID(ctx, userID)
	if err2 != nil {
		slog.Error("failed to get accepter for friend accept notification", "user_id", userID, "error", err2)
	}
	if accepter != nil {
		if s.notificationService != nil {
			_, _ = s.notificationService.Create(ctx, &model.Notification{
				UserID: friendship.UserID,
				Type:   "friend_accepted",
				Title:  "好友请求已通过",
				Body:   accepter.Username + " 接受了你的好友请求",
				RefID:  friendship.ID,
			})
		}
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeFriendAccepted,
			UserID: friendship.UserID,
			Data: map[string]any{
				"friend": accepter.ToInfo(),
			},
		})
	}

	return nil
}

// RejectRequest rejects (deletes) a pending friend request. Only the receiver can reject.
func (s *FriendService) RejectRequest(ctx context.Context, userID, requestID int64) error {
	friendship, err := s.friendRepo.GetByID(ctx, requestID)
	if err != nil {
		return err
	}
	if friendship == nil || friendship.Status != "pending" {
		return ErrFriendReqNotFound
	}
	if friendship.FriendID != userID {
		return ErrNoPermission
	}

	if err := s.friendRepo.Delete(ctx, requestID); err != nil {
		return err
	}

	// WS notification: notify requester that their request was rejected
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeFriendRejected,
		UserID: friendship.UserID,
		Data: map[string]any{
			"request_id": requestID,
		},
	})

	return nil
}

// DeleteFriend removes an accepted friendship. Either party can delete.
func (s *FriendService) DeleteFriend(ctx context.Context, userID, friendshipID int64) error {
	friendship, err := s.friendRepo.GetByID(ctx, friendshipID)
	if err != nil {
		return err
	}
	if friendship == nil {
		return ErrNotFriends
	}

	// Ensure the user is part of this friendship
	var otherUserID int64
	if friendship.UserID == userID {
		otherUserID = friendship.FriendID
	} else if friendship.FriendID == userID {
		otherUserID = friendship.UserID
	} else {
		return ErrNoPermission
	}

	if err := s.friendRepo.Delete(ctx, friendshipID); err != nil {
		return err
	}

	// WS notification: notify the other party
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeFriendRemoved,
		UserID: otherUserID,
		Data: map[string]any{
			"user_id": userID,
		},
	})

	return nil
}

// ListFriends returns all accepted friends for a user.
func (s *FriendService) ListFriends(ctx context.Context, userID int64) ([]*repository.FriendInfo, error) {
	return s.friendRepo.ListFriends(ctx, userID)
}

// ListPendingRequests returns all pending friend requests received by a user.
func (s *FriendService) ListPendingRequests(ctx context.Context, userID int64) ([]*repository.FriendRequestInfo, error) {
	return s.friendRepo.ListPendingRequests(ctx, userID)
}

// SearchUsers searches for users by username prefix.
func (s *FriendService) SearchUsers(ctx context.Context, query string, excludeUserID int64) ([]*model.UserInfo, error) {
	if len([]rune(query)) > 64 {
		query = string([]rune(query)[:64])
	}
	return s.friendRepo.SearchUsers(ctx, query, excludeUserID, 20)
}
