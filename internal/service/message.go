package service

import (
	"context"
	"time"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/security"
	"github.com/teamsphere/server/internal/ws"
)

const recallTimeLimit = 2 * time.Minute

type MessageService struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
	userRepo    repository.UserRepository
	hub         *ws.Hub
}

func NewMessageService(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository, userRepo repository.UserRepository, hub *ws.Hub) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
		userRepo:    userRepo,
		hub:         hub,
	}
}

// RecallRoomMessage recalls a room message.
// Permission rules:
//   - Sender can recall within 2 minutes
//   - Room admin/owner can recall any message (no time limit)
//   - System admin/owner can recall any message (no time limit)
func (s *MessageService) RecallRoomMessage(ctx context.Context, userID int64, systemRole string, roomID, msgID int64) error {
	msg, err := s.messageRepo.GetByID(ctx, msgID)
	if err != nil {
		return err
	}
	if msg == nil {
		return ErrMessageNotFound
	}
	if msg.RoomID != roomID {
		return ErrMessageNotFound
	}
	if msg.DeletedAt != nil {
		return ErrAlreadyRecalled
	}

	// Check permission
	allowed, err := s.canRecallRoomMessage(ctx, userID, systemRole, roomID, msg)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrRecallForbidden
	}

	if err := s.messageRepo.SoftDelete(ctx, msgID); err != nil {
		return err
	}

	// Broadcast msg_recalled to all room viewers
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeMsgRecalled,
		RoomID: roomID,
		Data: map[string]any{
			"msg_id":      msgID,
			"room_id":     roomID,
			"recalled_by": userID,
		},
	})

	return nil
}

// RecallDM recalls a direct message.
// Permission rules:
//   - Sender can recall within 2 minutes
//   - System admin/owner can recall any DM (no time limit)
func (s *MessageService) RecallDM(ctx context.Context, userID int64, systemRole string, msgID int64) error {
	dm, err := s.messageRepo.GetDMByID(ctx, msgID)
	if err != nil {
		return err
	}
	if dm == nil {
		return ErrMessageNotFound
	}
	if dm.DeletedAt != nil {
		return ErrAlreadyRecalled
	}

	// Check permission
	isSender := dm.SenderID == userID
	isSysAdmin := systemRoleLevel(systemRole) >= 2

	if isSender {
		if time.Since(dm.CreatedAt) > recallTimeLimit {
			if !isSysAdmin {
				return ErrRecallTimeout
			}
		}
	} else if !isSysAdmin {
		return ErrRecallForbidden
	}

	if err := s.messageRepo.SoftDeleteDM(ctx, msgID); err != nil {
		return err
	}

	// Notify both sender and receiver
	data := map[string]any{
		"msg_id":      msgID,
		"recalled_by": userID,
	}

	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeDMRecalled,
		UserID: dm.SenderID,
		Data:   data,
	})
	if dm.ReceiverID != dm.SenderID {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeDMRecalled,
			UserID: dm.ReceiverID,
			Data:   data,
		})
	}

	return nil
}

func (s *MessageService) canRecallRoomMessage(ctx context.Context, userID int64, systemRole string, roomID int64, msg *model.Message) (bool, error) {
	// System admin+ can always recall
	if systemRoleLevel(systemRole) >= 2 {
		return true, nil
	}

	// Room admin/owner can recall any message
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return false, err
	}
	if member != nil && roleLevel(member.Role) >= roleLevel("admin") {
		return true, nil
	}

	// Sender can recall within time limit
	if msg.UserID != nil && *msg.UserID == userID {
		if time.Since(msg.CreatedAt) <= recallTimeLimit {
			return true, nil
		}
		return false, ErrRecallTimeout
	}

	return false, nil
}

// EditRoomMessage edits a room message's content.
// Only the original sender can edit their own messages.
func (s *MessageService) EditRoomMessage(ctx context.Context, userID int64, roomID, msgID int64, newContent string) error {
	msg, err := s.messageRepo.GetByID(ctx, msgID)
	if err != nil {
		return err
	}
	if msg == nil {
		return ErrMessageNotFound
	}
	if msg.RoomID != roomID {
		return ErrMessageNotFound
	}
	if msg.DeletedAt != nil {
		return ErrAlreadyRecalled
	}
	// Only the sender can edit
	if msg.UserID == nil || *msg.UserID != userID {
		return ErrEditForbidden
	}

	safe := security.SanitizeMessageContent(newContent)
	updatedAt, err := s.messageRepo.UpdateContent(ctx, msgID, safe)
	if err != nil {
		return err
	}

	// Broadcast msg_edited to all room viewers
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeMsgEdited,
		RoomID: roomID,
		Data: map[string]any{
			"msg_id":     msgID,
			"room_id":    roomID,
			"user_id":    userID,
			"content":    safe,
			"updated_at": updatedAt,
		},
	})

	return nil
}

// EditDM edits a direct message's content.
// Only the original sender can edit their own messages.
func (s *MessageService) EditDM(ctx context.Context, userID int64, msgID int64, newContent string) error {
	dm, err := s.messageRepo.GetDMByID(ctx, msgID)
	if err != nil {
		return err
	}
	if dm == nil {
		return ErrMessageNotFound
	}
	if dm.DeletedAt != nil {
		return ErrAlreadyRecalled
	}
	// Only the sender can edit
	if dm.SenderID != userID {
		return ErrEditForbidden
	}

	safe := security.SanitizeMessageContent(newContent)
	updatedAt, err := s.messageRepo.UpdateDMContent(ctx, msgID, safe)
	if err != nil {
		return err
	}

	// Notify both sender and receiver
	data := map[string]any{
		"msg_id":     msgID,
		"user_id":    userID,
		"content":    safe,
		"updated_at": updatedAt,
	}

	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeDMEdited,
		UserID: dm.SenderID,
		Data:   data,
	})
	if dm.ReceiverID != dm.SenderID {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeDMEdited,
			UserID: dm.ReceiverID,
			Data:   data,
		})
	}

	return nil
}

// MarkDMRead marks direct messages from peer as read and notifies the sender.
func (s *MessageService) MarkDMRead(ctx context.Context, userID, peerID, lastReadMsgID int64) (int64, *time.Time, error) {
	if peerID <= 0 {
		return 0, nil, ErrMessageNotFound
	}
	count, readAt, err := s.messageRepo.MarkDMRead(ctx, userID, peerID, lastReadMsgID)
	if err != nil {
		return 0, nil, err
	}
	if count > 0 {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeDMRead,
			UserID: peerID,
			Data: map[string]any{
				"peer_id":          userID,
				"last_read_msg_id": lastReadMsgID,
				"read_at":          readAt,
			},
		})
	}
	return count, readAt, nil
}

// ForwardMessage forwards a message (room or dm) to a target room or user.
// If targetRoomID > 0, sends to room; otherwise to targetUserID.
func (s *MessageService) ForwardMessage(ctx context.Context, userID int64, sourceType string, sourceMsgID, targetRoomID, targetUserID int64, comment string) (*model.Message, *model.DirectMessage, error) {
	var forward *model.ForwardInfo

	switch sourceType {
	case "room":
		msg, err := s.messageRepo.GetByID(ctx, sourceMsgID)
		if err != nil {
			return nil, nil, err
		}
		if msg == nil {
			return nil, nil, ErrMessageNotFound
		}
		member, err := s.roomRepo.GetMember(ctx, msg.RoomID, userID)
		if err != nil {
			return nil, nil, err
		}
		if member == nil {
			return nil, nil, ErrNotRoomMember
		}
		originUser, _ := s.userRepo.GetByID(ctx, safeInt64(msg.UserID))
		forward = buildForwardInfo("room", sourceMsgID, msg.MsgType, msg.Content, msg.DeletedAt != nil, originUser)
	case "dm":
		dm, err := s.messageRepo.GetDMByID(ctx, sourceMsgID)
		if err != nil {
			return nil, nil, err
		}
		if dm == nil {
			return nil, nil, ErrMessageNotFound
		}
		if dm.SenderID != userID && dm.ReceiverID != userID {
			return nil, nil, ErrNoPermission
		}
		originUser, _ := s.userRepo.GetByID(ctx, dm.SenderID)
		forward = buildForwardInfo("dm", sourceMsgID, dm.MsgType, dm.Content, dm.DeletedAt != nil, originUser)
	default:
		return nil, nil, ErrInvalidParams
	}

	senderInfo := model.UserInfo{ID: userID}
	if u, _ := s.userRepo.GetByID(ctx, userID); u != nil {
		senderInfo.Username = u.Username
		senderInfo.AvatarURL = u.AvatarURL
	}

	if targetRoomID > 0 {
		safeComment := security.SanitizeMessageContent(comment)
		member, err := s.roomRepo.GetMember(ctx, targetRoomID, userID)
		if err != nil {
			return nil, nil, err
		}
		if member == nil {
			return nil, nil, ErrNotRoomMember
		}
		dbMsg, err := s.messageRepo.Create(ctx, safeComment, userID, targetRoomID, "forward", nil, nil, nil, nil, nil, forward)
		if err != nil {
			return nil, nil, err
		}
		chatMsg := &ws.ChatBroadcast{
			ID:          dbMsg.ID,
			Content:     dbMsg.Content,
			MsgType:     dbMsg.MsgType,
			ForwardMeta: dbMsg.ForwardMeta,
			User:        senderInfo,
			RoomID:    targetRoomID,
			Mentions:  dbMsg.Mentions,
			ReplyTo:   nil,
			CreatedAt: dbMsg.CreatedAt,
		}
		s.hub.SendAction(&ws.Action{Type: ws.TypeChat, RoomID: targetRoomID, Data: chatMsg})
		return dbMsg, nil, nil
	}

	if targetUserID > 0 {
		safeComment := security.SanitizeMessageContent(comment)
		dbDM, err := s.messageRepo.CreateDM(ctx, safeComment, userID, targetUserID, "forward", nil, nil, nil, nil, forward)
		if err != nil {
			return nil, nil, err
		}
		dmBroadcast := &ws.DMBroadcast{
			ID:          dbDM.ID,
			Content:     dbDM.Content,
			MsgType:     dbDM.MsgType,
			ForwardMeta: dbDM.ForwardMeta,
			User:        senderInfo,
			CreatedAt: dbDM.CreatedAt,
		}
		s.hub.SendAction(&ws.Action{Type: ws.TypeDM, UserID: targetUserID, Data: dmBroadcast})
		s.hub.SendAction(&ws.Action{Type: ws.TypeDM, UserID: userID, Data: dmBroadcast})
		return nil, dbDM, nil
	}

	return nil, nil, ErrInvalidParams
}

func buildForwardInfo(sourceType string, sourceID int64, msgType, content string, deleted bool, user *model.User) *model.ForwardInfo {
	fi := &model.ForwardInfo{
		Type:      sourceType,
		ID:        sourceID,
		MsgType:   msgType,
		IsDeleted: deleted,
	}
	if !deleted {
		fi.Content = truncateForwardContent(content, 200)
	}
	if user != nil {
		fi.User = model.UserInfo{ID: user.ID, Username: user.Username, AvatarURL: user.AvatarURL}
	} else {
		fi.User = model.UserInfo{ID: 0, Username: "Unknown"}
	}
	return fi
}

func truncateForwardContent(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func safeInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}


// ListThreadMessages returns messages in a thread for a room.
func (s *MessageService) ListThreadMessages(ctx context.Context, userID, roomID, rootMsgID int64, beforeID, afterID int64, limit int) ([]*repository.MessageWithUser, error) {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotRoomMember
	}
	root, err := s.messageRepo.GetByID(ctx, rootMsgID)
	if err != nil {
		return nil, err
	}
	if root == nil || root.RoomID != roomID {
		return nil, ErrMessageNotFound
	}
	return s.messageRepo.ListThreadByRoom(ctx, roomID, rootMsgID, beforeID, afterID, limit)
}

// PinRoomMessage pins a message in a room (admin/owner only).
func (s *MessageService) PinRoomMessage(ctx context.Context, userID int64, systemRole string, roomID, msgID int64) error {
	msg, err := s.messageRepo.GetByID(ctx, msgID)
	if err != nil {
		return err
	}
	if msg == nil || msg.RoomID != roomID {
		return ErrMessageNotFound
	}
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if systemRoleLevel(systemRole) < 2 {
		if member == nil || roleLevel(member.Role) < roleLevel("admin") {
			return ErrNoPermission
		}
	}
	if err := s.messageRepo.PinMessage(ctx, roomID, msgID, userID); err != nil {
		return err
	}
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeMsgPinned,
		RoomID: roomID,
		Data: map[string]any{
			"msg_id":    msgID,
			"room_id":   roomID,
			"pinned_by": userID,
		},
	})
	return nil
}

// UnpinRoomMessage unpins a message in a room (admin/owner only).
func (s *MessageService) UnpinRoomMessage(ctx context.Context, userID int64, systemRole string, roomID, msgID int64) error {
	msg, err := s.messageRepo.GetByID(ctx, msgID)
	if err != nil {
		return err
	}
	if msg == nil || msg.RoomID != roomID {
		return ErrMessageNotFound
	}
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return err
	}
	if systemRoleLevel(systemRole) < 2 {
		if member == nil || roleLevel(member.Role) < roleLevel("admin") {
			return ErrNoPermission
		}
	}
	if err := s.messageRepo.UnpinMessage(ctx, roomID, msgID); err != nil {
		return err
	}
	s.hub.SendAction(&ws.Action{
		Type:   ws.TypeMsgUnpinned,
		RoomID: roomID,
		Data: map[string]any{
			"msg_id":      msgID,
			"room_id":     roomID,
			"unpinned_by": userID,
		},
	})
	return nil
}

// ListPinnedMessages returns pinned messages for a room.
func (s *MessageService) ListPinnedMessages(ctx context.Context, userID, roomID int64) ([]*repository.MessageWithUser, error) {
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotRoomMember
	}
	return s.messageRepo.ListPinnedMessages(ctx, roomID)
}

// BatchRecallRoomMessages recalls multiple room messages (admin/owner only).
func (s *MessageService) BatchRecallRoomMessages(ctx context.Context, userID int64, systemRole string, roomID int64, msgIDs []int64) ([]int64, error) {
	if roomID <= 0 {
		return nil, ErrRoomNotFound
	}

	// Permission: room admin/owner or system admin/owner
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	isSysAdmin := systemRoleLevel(systemRole) >= 2
	if !isSysAdmin {
		if member == nil || roleLevel(member.Role) < roleLevel("admin") {
			return nil, ErrNoPermission
		}
	}

	ids, err := s.messageRepo.SoftDeleteBatch(ctx, roomID, msgIDs)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		s.hub.SendAction(&ws.Action{
			Type:   ws.TypeMsgRecalled,
			RoomID: roomID,
			Data: map[string]any{
				"msg_id":      id,
				"room_id":     roomID,
				"recalled_by": userID,
			},
		})
	}

	return ids, nil
}

// SearchMessages searches room messages with permission check.
func (s *MessageService) SearchMessages(ctx context.Context, userID int64, query string, roomID, senderID int64, from, to *time.Time, limit int) ([]*repository.MessageWithUser, error) {
	if roomID <= 0 {
		return nil, ErrRoomNotFound
	}
	member, err := s.roomRepo.GetMember(ctx, roomID, userID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrNotRoomMember
	}
	return s.messageRepo.SearchMessages(ctx, userID, query, roomID, senderID, from, to, limit)
}
