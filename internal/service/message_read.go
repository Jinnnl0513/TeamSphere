package service

import (
	"context"
	"fmt"

	"github.com/teamsphere/server/internal/repository"
)

type MessageReadService struct {
	readRepo repository.MessageReadRepository
	roomRepo repository.RoomRepository
}

func NewMessageReadService(readRepo repository.MessageReadRepository, roomRepo repository.RoomRepository) *MessageReadService {
	return &MessageReadService{readRepo: readRepo, roomRepo: roomRepo}
}

func (s *MessageReadService) MarkRead(ctx context.Context, userID, roomID, lastReadMsgID int64) error {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrNotRoomMember
	}

	if lastReadMsgID <= 0 {
		_, err := s.readRepo.MarkReadAtLatest(ctx, userID, roomID)
		return err
	}
	_, err = s.readRepo.Upsert(ctx, userID, roomID, lastReadMsgID)
	return err
}

func (s *MessageReadService) UnreadCount(ctx context.Context, userID, roomID int64) (int64, error) {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return 0, err
	}
	if member == nil {
		return 0, ErrNotRoomMember
	}
	lastRead, err := s.readRepo.GetLastReadID(ctx, userID, roomID)
	if err != nil {
		return 0, err
	}
	count, err := s.readRepo.GetUnreadCount(ctx, userID, roomID, lastRead)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *MessageReadService) BatchUnreadCounts(ctx context.Context, userID int64, roomIDs []int64) (map[int64]int64, error) {
	out := make(map[int64]int64, len(roomIDs))
	for _, roomID := range roomIDs {
		count, err := s.UnreadCount(ctx, userID, roomID)
		if err != nil {
			return nil, fmt.Errorf("unread count for room %d: %w", roomID, err)
		}
		out[roomID] = count
	}
	return out, nil
}
