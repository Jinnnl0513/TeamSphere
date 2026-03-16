package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

type DMHandler struct {
	messageRepo    repository.MessageRepository
	friendshipRepo repository.FriendshipRepository
	messageService *service.MessageService
}

func NewDMHandler(messageRepo repository.MessageRepository, friendshipRepo repository.FriendshipRepository, messageService *service.MessageService) *DMHandler {
	return &DMHandler{messageRepo: messageRepo, friendshipRepo: friendshipRepo, messageService: messageService}
}

// ListMessages handles GET /dm/:user_id/messages with cursor pagination.
func (h *DMHandler) ListMessages(c *gin.Context) {
	peerID, ok := parseIDParam(c, "user_id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")

	// Verify friendship
	friends, err := h.friendshipRepo.AreFriends(c.Request.Context(), userID, peerID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	if !friends {
		Error(c, http.StatusForbidden, 40301, "没有权限")
		return
	}

	beforeID := parseOptionalInt64(c.Query("before_id"))
	afterID := parseOptionalInt64(c.Query("after_id"))
	limit := parseOptionalInt(c.Query("limit"), 50)

	if beforeID > 0 && afterID > 0 {
		Error(c, http.StatusBadRequest, 40001, "before_id 鍜?after_id 涓嶈兘鍚屾椂浣跨敤")
		return
	}

	messages, err := h.messageRepo.ListDMs(c.Request.Context(), userID, peerID, beforeID, afterID, limit)
	if err != nil {
		slog.Error("failed to list dms", "error", err)
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	if messages == nil {
		messages = []*repository.DMWithUser{}
	}
	Success(c, messages)
}

// ListConversations handles GET /dm/conversations.
func (h *DMHandler) ListConversations(c *gin.Context) {
	userID := c.GetInt64("user_id")

	convos, err := h.messageRepo.ListConversations(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list conversations", "error", err)
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	if convos == nil {
		convos = []*repository.Conversation{}
	}
	Success(c, convos)
}

// RecallDM handles DELETE /dm/messages/:msg_id.
func (h *DMHandler) RecallDM(c *gin.Context) {
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.messageService.RecallDM(c.Request.Context(), userID, c.GetString("role"), msgID); err != nil {
		handleDMError(c, err)
		return
	}
	Success(c, nil)
}

func handleDMError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrMessageNotFound):
		Error(c, http.StatusNotFound, 40401, "资源不存在")
	case errors.Is(err, service.ErrRecallTimeout):
		Error(c, http.StatusForbidden, 40304, "宸茶秴杩囨秷鎭彲鎾ゅ洖鏃堕棿")
	case errors.Is(err, service.ErrRecallForbidden):
		Error(c, http.StatusForbidden, 40305, "没有权限")
	case errors.Is(err, service.ErrAlreadyRecalled):
		Error(c, http.StatusConflict, 40901, "娑堟伅宸茶鎾ゅ洖")
	case errors.Is(err, service.ErrEditForbidden):
		Error(c, http.StatusForbidden, 40301, "鍙湁鍙戦€佽€呭彲浠ョ紪杈戞娑堟伅")
	default:
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
	}
}

type editDMRequest struct {
	Content string `json:"content" binding:"required"`
}

// EditDM handles PUT /dm/messages/:msg_id.
func (h *DMHandler) EditDM(c *gin.Context) {
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}

	var req editDMRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "璇锋眰鍙傛暟閿欒")
		return
	}
	if len([]rune(req.Content)) > 2000 {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.messageService.EditDM(c.Request.Context(), userID, msgID, req.Content); err != nil {
		handleDMError(c, err)
		return
	}
	Success(c, nil)
}

type markDMReadRequest struct {
	LastReadMsgID int64 `json:"last_read_msg_id"`
}

// MarkRead handles POST /dm/:user_id/read.
func (h *DMHandler) MarkRead(c *gin.Context) {
	peerID, ok := parseIDParam(c, "user_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")

	// Verify friendship
	friends, err := h.friendshipRepo.AreFriends(c.Request.Context(), userID, peerID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	if !friends {
		Error(c, http.StatusForbidden, 40301, "没有权限")
		return
	}

	var req markDMReadRequest
	_ = c.ShouldBindJSON(&req)

	count, readAt, err := h.messageService.MarkDMRead(c.Request.Context(), userID, peerID, req.LastReadMsgID)
	if err != nil {
		handleDMError(c, err)
		return
	}
	Success(c, gin.H{"read_count": count, "read_at": readAt})
}
