package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Members

// ListMembers handles GET /rooms/:id/members.
func (h *RoomHandler) ListMembers(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	members, err := h.roomService.ListMembers(c.Request.Context(), userID, roomID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, members)
}

type updateMemberRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateMemberRole handles PUT /rooms/:id/members/:user_id.
func (h *RoomHandler) UpdateMemberRole(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	targetUserID, ok := parseIDParam(c, "user_id")
	if !ok {
		return
	}

	var req updateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.UpdateMemberRole(c.Request.Context(), userID, c.GetString("role"), roomID, targetUserID, req.Role); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_member_role_change", "room_member", targetUserID, map[string]any{"room_id": roomID, "role": req.Role})
	Success(c, nil)
}

// KickMember handles DELETE /rooms/:id/members/:user_id.
func (h *RoomHandler) KickMember(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	targetUserID, ok := parseIDParam(c, "user_id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.KickMember(c.Request.Context(), userID, c.GetString("role"), roomID, targetUserID); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_member_kick", "room_member", targetUserID, map[string]any{"room_id": roomID})
	Success(c, nil)
}

type muteRequest struct {
	Duration int `json:"duration" binding:"required,min=1"`
}

// MuteMember handles POST /rooms/:id/members/:user_id/mute.
func (h *RoomHandler) MuteMember(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	targetUserID, ok := parseIDParam(c, "user_id")
	if !ok {
		return
	}

	var req muteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "绂佽█鏃堕暱(鍒嗛挓)涓哄繀濉」")
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.MuteMember(c.Request.Context(), userID, c.GetString("role"), roomID, targetUserID, req.Duration); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_member_mute", "room_member", targetUserID, map[string]any{"room_id": roomID, "duration": req.Duration})
	Success(c, nil)
}

// UnmuteMember handles DELETE /rooms/:id/members/:user_id/mute.
func (h *RoomHandler) UnmuteMember(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	targetUserID, ok := parseIDParam(c, "user_id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.UnmuteMember(c.Request.Context(), userID, c.GetString("role"), roomID, targetUserID); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_member_unmute", "room_member", targetUserID, map[string]any{"room_id": roomID})
	Success(c, nil)
}

type transferRequest struct {
	NewOwnerID int64 `json:"new_owner_id" binding:"required"`
}

// TransferOwner handles PUT /rooms/:id/transfer.
func (h *RoomHandler) TransferOwner(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req transferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "鏂扮殑缇や富ID涓哄繀濉」")
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.TransferOwner(c.Request.Context(), userID, roomID, req.NewOwnerID); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}
