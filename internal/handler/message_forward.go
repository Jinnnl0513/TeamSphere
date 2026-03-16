package handler

import (
	"errors"
	"net/http"

	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageRepo    repository.MessageRepository
	roomRepo       repository.RoomRepository
	friendshipRepo repository.FriendshipRepository
	messageService *service.MessageService
}

func NewMessageHandler(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository, friendshipRepo repository.FriendshipRepository, messageService *service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageRepo:    messageRepo,
		roomRepo:       roomRepo,
		friendshipRepo: friendshipRepo,
		messageService: messageService,
	}
}

type forwardMessageRequest struct {
	MessageType  string `json:"message_type" binding:"required,oneof=room dm"`
	TargetRoomID int64  `json:"target_room_id"`
	TargetUserID int64  `json:"target_user_id"`
	Comment      string `json:"comment"`
}

// ForwardMessage handles POST /messages/:msg_id/forward.
func (h *MessageHandler) ForwardMessage(c *gin.Context) {
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")

	var req forwardMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if (req.TargetRoomID > 0 && req.TargetUserID > 0) || (req.TargetRoomID == 0 && req.TargetUserID == 0) {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if len([]rune(req.Comment)) > 2000 {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	switch req.MessageType {
	case "room":
		msg, err := h.messageRepo.GetByID(c.Request.Context(), msgID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
			return
		}
		if msg == nil {
			Error(c, http.StatusNotFound, 40401, "资源不存在")
			return
		}
		member, err := h.roomRepo.GetMember(c.Request.Context(), msg.RoomID, userID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
			return
		}
		if member == nil {
			Error(c, http.StatusForbidden, 40301, "涓嶆槸鎴块棿鎴愬憳")
			return
		}
	case "dm":
		dm, err := h.messageRepo.GetDMByID(c.Request.Context(), msgID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
			return
		}
		if dm == nil {
			Error(c, http.StatusNotFound, 40401, "资源不存在")
			return
		}
		if dm.SenderID != userID && dm.ReceiverID != userID {
			Error(c, http.StatusForbidden, 40301, "没有权限")
			return
		}
	default:
		Error(c, http.StatusBadRequest, 40001, "message_type 蹇呴』鏄?room 鎴?dm")
		return
	}

	if req.TargetRoomID > 0 {
		member, err := h.roomRepo.GetMember(c.Request.Context(), req.TargetRoomID, userID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
			return
		}
		if member == nil {
			Error(c, http.StatusForbidden, 40301, "涓嶆槸鐩爣鎴块棿鎴愬憳")
			return
		}
	} else if req.TargetUserID > 0 {
		friends, err := h.friendshipRepo.AreFriends(c.Request.Context(), userID, req.TargetUserID)
		if err != nil {
			Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
			return
		}
		if !friends {
			Error(c, http.StatusForbidden, 40301, "没有权限")
			return
		}
	}

	msg, dm, err := h.messageService.ForwardMessage(c.Request.Context(), userID, req.MessageType, msgID, req.TargetRoomID, req.TargetUserID, req.Comment)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidParams):
			Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		case errors.Is(err, service.ErrNotRoomMember):
			Error(c, http.StatusForbidden, 40301, "涓嶆槸鎴块棿鎴愬憳")
		case errors.Is(err, service.ErrNoPermission):
			Error(c, http.StatusForbidden, 40301, "娌℃湁鏉冮檺")
		case errors.Is(err, service.ErrMessageNotFound):
			Error(c, http.StatusNotFound, 40401, "资源不存在")
		default:
			Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		}
		return
	}
	if msg != nil {
		Success(c, msg)
		return
	}
	Success(c, dm)
}
