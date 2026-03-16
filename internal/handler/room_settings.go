package handler

import (
	"net/http"

	"github.com/teamsphere/server/internal/model"
	"github.com/gin-gonic/gin"
)

type updateRoomSettingsRequest struct {
	Settings model.RoomSettings `json:"settings" binding:"required"`
}

type updateRoomPermissionsRequest struct {
	Permissions []model.RoomRolePermission `json:"permissions" binding:"required"`
}

type statsQuery struct {
	Days int `form:"days"`
}

// GetRoomSettings handles GET /rooms/:id/settings.
func (h *RoomHandler) GetRoomSettings(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	settings, err := h.settingsService.GetSettings(c.Request.Context(), roomID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, settings)
}

// UpdateRoomSettings handles PUT /rooms/:id/settings.
func (h *RoomHandler) UpdateRoomSettings(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateRoomSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userID := c.GetInt64("user_id")
	systemRole := c.GetString("role")
	updated, err := h.settingsService.UpdateSettings(c.Request.Context(), userID, systemRole, roomID, &req.Settings)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, updated)
}

// GetRoomPermissions handles GET /rooms/:id/permissions.
func (h *RoomHandler) GetRoomPermissions(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	perms, err := h.settingsService.ListPermissions(c.Request.Context(), roomID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, perms)
}

// UpdateRoomPermissions handles PUT /rooms/:id/permissions.
func (h *RoomHandler) UpdateRoomPermissions(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var req updateRoomPermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userID := c.GetInt64("user_id")
	systemRole := c.GetString("role")
	perms := make([]*model.RoomRolePermission, 0, len(req.Permissions))
	for i := range req.Permissions {
		p := req.Permissions[i]
		perms = append(perms, &p)
	}
	if err := h.settingsService.UpdatePermissions(c.Request.Context(), userID, systemRole, roomID, perms); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}

// ListJoinRequests handles GET /rooms/:id/join-requests.
func (h *RoomHandler) ListJoinRequests(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	systemRole := c.GetString("role")
	list, err := h.settingsService.ListJoinRequests(c.Request.Context(), userID, systemRole, roomID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, list)
}

// ApproveJoinRequest handles POST /rooms/:id/join-requests/:req_id/approve.
func (h *RoomHandler) ApproveJoinRequest(c *gin.Context) {
	_, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	reqID, ok := parseIDParam(c, "req_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	systemRole := c.GetString("role")
	if err := h.settingsService.ApproveJoinRequest(c.Request.Context(), userID, systemRole, reqID); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}

// RejectJoinRequest handles POST /rooms/:id/join-requests/:req_id/reject.
func (h *RoomHandler) RejectJoinRequest(c *gin.Context) {
	_, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	reqID, ok := parseIDParam(c, "req_id")
	if !ok {
		return
	}
	userID := c.GetInt64("user_id")
	systemRole := c.GetString("role")
	if err := h.settingsService.RejectJoinRequest(c.Request.Context(), userID, systemRole, reqID); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}

// GetRoomStatsSummary handles GET /rooms/:id/stats/summary.
func (h *RoomHandler) GetRoomStatsSummary(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	var q statsQuery
	_ = c.ShouldBindQuery(&q)
	userID := c.GetInt64("user_id")
	systemRole := c.GetString("role")
	res, err := h.settingsService.GetStatsSummary(c.Request.Context(), userID, systemRole, roomID, q.Days)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, res)
}
