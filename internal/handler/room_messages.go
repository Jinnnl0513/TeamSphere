package handler

import (
	"net/http"

	"github.com/teamsphere/server/internal/repository"
	"github.com/gin-gonic/gin"
)

type markReadRequest struct {
	LastReadMsgID int64 `json:"last_read_msg_id"`
}

// MarkRead handles POST /rooms/:id/read.
func (h *RoomHandler) MarkRead(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req markReadRequest
	_ = c.ShouldBindJSON(&req)
	userID := c.GetInt64("user_id")
	if err := h.readService.MarkRead(c.Request.Context(), userID, roomID, req.LastReadMsgID); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}

// UnreadCount handles GET /rooms/:id/unread-count.
func (h *RoomHandler) UnreadCount(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	count, err := h.readService.UnreadCount(c.Request.Context(), userID, roomID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, gin.H{"unread_count": count})
}

// ListPinnedMessages handles GET /rooms/:id/pinned-messages.
func (h *RoomHandler) ListPinnedMessages(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	messages, err := h.messageService.ListPinnedMessages(c.Request.Context(), userID, roomID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	if messages == nil {
		messages = []*repository.MessageWithUser{}
	}
	Success(c, messages)
}

type batchDeleteMessagesRequest struct {
	MsgIDs []int64 `json:"msg_ids"`
}

// BatchDeleteRoomMessages handles DELETE /rooms/:id/messages/batch.
func (h *RoomHandler) BatchDeleteRoomMessages(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req batchDeleteMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	if len(req.MsgIDs) == 0 {
		Error(c, http.StatusBadRequest, 40001, "msg_ids 涓嶈兘涓虹┖")
		return
	}
	if len(req.MsgIDs) > 200 {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}
	for _, id := range req.MsgIDs {
		if id <= 0 {
			Error(c, http.StatusBadRequest, 40001, "msg_ids 蹇呴』涓烘鏁存暟")
			return
		}
	}

	userID := c.GetInt64("user_id")
	ids, err := h.messageService.BatchRecallRoomMessages(c.Request.Context(), userID, c.GetString("role"), roomID, req.MsgIDs)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_message_batch_delete", "message", roomID, map[string]any{"room_id": roomID, "msg_ids": ids})
	Success(c, gin.H{"deleted_ids": ids})
}

// PinRoomMessage handles POST /rooms/:id/messages/:msg_id/pin.
func (h *RoomHandler) PinRoomMessage(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.messageService.PinRoomMessage(c.Request.Context(), userID, c.GetString("role"), roomID, msgID); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_message_pin", "message", msgID, map[string]any{"room_id": roomID})
	Success(c, nil)
}

// UnpinRoomMessage handles DELETE /rooms/:id/messages/:msg_id/pin.
func (h *RoomHandler) UnpinRoomMessage(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.messageService.UnpinRoomMessage(c.Request.Context(), userID, c.GetString("role"), roomID, msgID); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_message_unpin", "message", msgID, map[string]any{"room_id": roomID})
	Success(c, nil)
}

// RecallRoomMessage handles DELETE /rooms/:id/messages/:msg_id.
func (h *RoomHandler) RecallRoomMessage(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.messageService.RecallRoomMessage(c.Request.Context(), userID, c.GetString("role"), roomID, msgID); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_message_recall", "message", msgID, map[string]any{"room_id": roomID})
	Success(c, nil)
}

type editMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// ListThreadMessages handles GET /rooms/:id/messages/:msg_id/thread.
func (h *RoomHandler) ListThreadMessages(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	beforeID := parseOptionalInt64(c.Query("before_id"))
	afterID := parseOptionalInt64(c.Query("after_id"))
	limit := parseOptionalInt(c.Query("limit"), 50)
	if beforeID > 0 && afterID > 0 {
		Error(c, http.StatusBadRequest, 40001, "before_id鍜宎fter_id涓嶈兘鍚屾椂浣跨敤")
		return
	}

	messages, err := h.messageService.ListThreadMessages(c.Request.Context(), userID, roomID, msgID, beforeID, afterID, limit)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	if messages == nil {
		messages = []*repository.MessageWithUser{}
	}
	Success(c, messages)
}

// EditRoomMessage handles PUT /rooms/:id/messages/:msg_id.
func (h *RoomHandler) EditRoomMessage(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	msgID, ok := parseIDParam(c, "msg_id")
	if !ok {
		return
	}

	var req editMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "鍐呭涓嶈兘涓虹┖")
		return
	}
	if len([]rune(req.Content)) > 2000 {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.messageService.EditRoomMessage(c.Request.Context(), userID, roomID, msgID, req.Content); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_message_edit", "message", msgID, map[string]any{"room_id": roomID})
	Success(c, nil)
}
