package ws

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/teamsphere/server/internal/model"
)

// broadcastToRoom sends a message to all clients viewing a room.
func (h *Hub) broadcastToRoom(roomID int64, msgType string, data any, skip *Client) {
	env, err := NewEnvelope(msgType, data)
	if err != nil {
		slog.Error("failed to create envelope", "type", msgType, "error", err)
		return
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return
	}

	if roomClients, ok := h.rooms[roomID]; ok {
		for c := range roomClients {
			if c == skip {
				continue
			}
			select {
			case c.send <- raw:
			default:
				slog.Warn("client send buffer full, message dropped",
					"user_id", c.UserID, "msg_type", msgType)
			}
		}
	}
}

// sendToClient sends a typed message to a single client.
func (h *Hub) sendToClient(c *Client, msgType string, data any) {
	env, err := NewEnvelope(msgType, data)
	if err != nil {
		return
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return
	}
	select {
	case c.send <- raw:
	default:
		slog.Warn("client send buffer full (sendToClient), message dropped",
			"user_id", c.UserID, "msg_type", msgType)
	}
}

// sendToUser sends a message to all connections of a user.
func (h *Hub) sendToUser(userID int64, msgType string, data any) {
	conns, ok := h.userIndex[userID]
	if !ok {
		return
	}
	env, err := NewEnvelope(msgType, data)
	if err != nil {
		return
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return
	}
	for c := range conns {
		select {
		case c.send <- raw:
		default:
			slog.Warn("client send buffer full (sendToUser), message dropped",
				"user_id", c.UserID, "msg_type", msgType)
		}
	}
}

func (h *Hub) sendMentionNotifications(ctx context.Context, sender *Client, roomID int64, dbMsg *model.Message) {
	room, err := h.roomRepo.GetByID(ctx, roomID)
	if err != nil || room == nil {
		return
	}

	contentPreview := dbMsg.Content
	if len([]rune(contentPreview)) > 100 {
		contentPreview = string([]rune(contentPreview)[:100]) + "..."
	}

	senderInfo := model.UserInfo{
		ID:        sender.UserID,
		Username:  sender.Username,
		AvatarURL: sender.AvatarURL,
	}

	roomViewers := h.rooms[roomID]

	notifData := map[string]any{
		"msg_id":          dbMsg.ID,
		"room_id":         roomID,
		"room_name":       room.Name,
		"user":            senderInfo,
		"content_preview": contentPreview,
	}

	hasAll := false
	for _, uid := range dbMsg.Mentions {
		if uid == 0 {
			hasAll = true
			break
		}
	}

	if hasAll {
		members, err := h.roomRepo.ListMemberUsernames(ctx, roomID)
		if err != nil {
			return
		}
		for _, uid := range members {
			if uid == sender.UserID {
				continue
			}
			if h.isUserViewingRoom(uid, roomViewers) {
				continue
			}
			h.createNotification(ctx, uid, "mention", "你被提及", contentPreview, dbMsg.ID)
			h.sendToUser(uid, TypeMentioned, notifData)
		}
	} else {
		for _, mentionedUID := range dbMsg.Mentions {
			if mentionedUID == sender.UserID {
				continue
			}
			if h.isUserViewingRoom(mentionedUID, roomViewers) {
				continue
			}
			h.createNotification(ctx, mentionedUID, "mention", "你被提及", contentPreview, dbMsg.ID)
			h.sendToUser(mentionedUID, TypeMentioned, notifData)
		}
	}
}

// isUserViewingRoom checks if a user has any client currently viewing the given room.
func (h *Hub) isUserViewingRoom(userID int64, roomViewers map[*Client]bool) bool {
	conns, ok := h.userIndex[userID]
	if !ok {
		return false
	}
	for c := range conns {
		if roomViewers[c] {
			return true
		}
	}
	return false
}

func (h *Hub) createNotification(ctx context.Context, userID int64, ntype, title, body string, refID int64) {
	if h.notificationRepo == nil {
		return
	}
	_, err := h.notificationRepo.Create(ctx, &model.Notification{
		UserID: userID,
		Type:   ntype,
		Title:  title,
		Body:   body,
		RefID:  refID,
	})
	if err != nil {
		slog.Warn("failed to persist notification", "error", err, "user_id", userID, "type", ntype)
	}
}
