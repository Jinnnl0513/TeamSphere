package handler

import (
	"net/http"
	"strings"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/ws"
	"github.com/gin-gonic/gin"
)

// ReactionHandler handles emoji reaction REST endpoints.
type ReactionHandler struct {
	reactionRepo *repository.ReactionRepo
	messageRepo  repository.MessageRepository
	hub          *ws.Hub
}

func NewReactionHandler(
	reactionRepo *repository.ReactionRepo,
	messageRepo repository.MessageRepository,
	hub *ws.Hub,
) *ReactionHandler {
	return &ReactionHandler{
		reactionRepo: reactionRepo,
		messageRepo:  messageRepo,
		hub:          hub,
	}
}

type addReactionReq struct {
	Emoji       string `json:"emoji"        binding:"required,max=64"`
	MessageType string `json:"message_type" binding:"required,oneof=room dm"`
	// RoomID is required for room messages so we can broadcast without extra DB lookup.
	RoomID int64 `json:"room_id"`
	// PeerUserID is required for dm messages so we can notify both parties.
	PeerUserID int64 `json:"peer_user_id"`
}

// AddReaction handles POST /messages/:msg_id/reactions
func (h *ReactionHandler) AddReaction(c *gin.Context) {
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")

	var req addReactionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "鍙傛暟閿欒: "+err.Error())
		return
	}
	req.Emoji = strings.TrimSpace(req.Emoji)

	added, err := h.reactionRepo.Add(c.Request.Context(), msgID, req.MessageType, userID, req.Emoji)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	summaries, err := h.reactionRepo.ListByMessage(c.Request.Context(), msgID, req.MessageType)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	payload := buildReactionBroadcast(msgID, req.MessageType, req.Emoji, userID, summaries)
	if added {
		act := &ws.Action{
			Type: ws.TypeReactionAdded,
			Data: payload,
		}
		if req.MessageType == "room" && req.RoomID != 0 {
			act.RoomID = req.RoomID
		} else if req.MessageType == "dm" && req.PeerUserID != 0 {
			// Notify peer; sender will see it via optimistic update
			act.UserID = req.PeerUserID
		}
		h.hub.SendAction(act)
	}

	Success(c, payload)
}

type removeReactionReq struct {
	MessageType string `json:"message_type" binding:"required,oneof=room dm"`
	RoomID      int64  `json:"room_id"`
	PeerUserID  int64  `json:"peer_user_id"`
}

// RemoveReaction handles DELETE /messages/:msg_id/reactions/*emoji
func (h *ReactionHandler) RemoveReaction(c *gin.Context) {
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}
	emoji := strings.TrimPrefix(c.Param("emoji"), "/")
	if emoji == "" {
		Error(c, http.StatusBadRequest, 40001, "emoji 涓嶈兘涓虹┖")
		return
	}
	userID := c.GetInt64("user_id")

	var req removeReactionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "鍙傛暟閿欒: "+err.Error())
		return
	}

	removed, err := h.reactionRepo.Remove(c.Request.Context(), msgID, req.MessageType, userID, emoji)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	summaries, err := h.reactionRepo.ListByMessage(c.Request.Context(), msgID, req.MessageType)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	payload := buildReactionBroadcast(msgID, req.MessageType, emoji, userID, summaries)
	if removed {
		act := &ws.Action{
			Type: ws.TypeReactionRemoved,
			Data: payload,
		}
		if req.MessageType == "room" && req.RoomID != 0 {
			act.RoomID = req.RoomID
		} else if req.MessageType == "dm" && req.PeerUserID != 0 {
			act.UserID = req.PeerUserID
		}
		h.hub.SendAction(act)
	}

	Success(c, payload)
}

// GetReactions handles GET /messages/:msg_id/reactions?type=room|dm
func (h *ReactionHandler) GetReactions(c *gin.Context) {
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}
	msgType := c.DefaultQuery("type", "room")
	if msgType != "room" && msgType != "dm" {
		Error(c, http.StatusBadRequest, 40001, "type 蹇呴』鏄?room 鎴?dm")
		return
	}

	summaries, err := h.reactionRepo.ListByMessage(c.Request.Context(), msgID, msgType)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	if summaries == nil {
		summaries = []*model.ReactionSummary{}
	}
	Success(c, summaries)
}

// buildReactionBroadcast constructs the WS broadcast payload for a reaction event.
func buildReactionBroadcast(messageID int64, messageType, emoji string, actorUserID int64, summaries []*model.ReactionSummary) ws.ReactionBroadcast {
	payload := ws.ReactionBroadcast{
		MessageID:   messageID,
		MessageType: messageType,
		Emoji:       emoji,
		UserID:      actorUserID,
		Count:       0,
		UserIDs:     []int64{},
	}
	for _, s := range summaries {
		if s.Emoji == emoji {
			payload.Count = s.Count
			payload.UserIDs = s.UserIDs
			if payload.UserIDs == nil {
				payload.UserIDs = []int64{}
			}
			break
		}
	}
	return payload
}
