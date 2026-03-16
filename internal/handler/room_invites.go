package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Invites

type inviteRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

// InviteFriend handles POST /rooms/:id/invite.
func (h *RoomHandler) InviteFriend(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req inviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "鐢ㄦ埛ID涓哄繀濉」")
		return
	}

	userID := c.GetInt64("user_id")
	invite, err := h.roomService.InviteFriend(c.Request.Context(), userID, c.GetString("role"), roomID, req.UserID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, invite)
}

// ListPendingInvites handles GET /rooms/invites.
func (h *RoomHandler) ListPendingInvites(c *gin.Context) {
	userID := c.GetInt64("user_id")
	invites, err := h.roomService.ListPendingInvites(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	Success(c, invites)
}

type respondInviteRequest struct {
	Action string `json:"action" binding:"required"` // "accept" or "decline"
}

// RespondInvite handles PUT /rooms/invites/:id.
func (h *RoomHandler) RespondInvite(c *gin.Context) {
	inviteID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req respondInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "鎿嶄綔绫诲瀷(action)蹇呭～锛屼笖鍙兘鏄?'accept' 鎴?'decline'")
		return
	}
	if req.Action != "accept" && req.Action != "decline" {
		Error(c, http.StatusBadRequest, 40001, "鎿嶄綔绫诲瀷(action)蹇呴』鏄?'accept' 鎴?'decline'")
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.RespondInvite(c.Request.Context(), userID, inviteID, req.Action == "accept"); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}
