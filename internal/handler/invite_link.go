package handler

import (
	"errors"
	"net/http"

	"github.com/teamsphere/server/internal/model"
	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

// InviteLinkHandler handles invite-link related HTTP endpoints.
type InviteLinkHandler struct {
	svc *service.InviteLinkService
}

func NewInviteLinkHandler(svc *service.InviteLinkService) *InviteLinkHandler {
	return &InviteLinkHandler{svc: svc}
}

type createInviteLinkRequest struct {
	MaxUses      int `json:"max_uses"`      // 0 = unlimited
	ExpiresHours int `json:"expires_hours"` // 0 = never
}

// CreateLink handles POST /rooms/:id/invite-links
func (h *InviteLinkHandler) CreateLink(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req createInviteLinkRequest
	// Body is optional; ignore bind errors (use defaults)
	_ = c.ShouldBindJSON(&req)

	userID := c.GetInt64("user_id")
	link, err := h.svc.CreateLink(c.Request.Context(), userID, roomID, req.MaxUses, req.ExpiresHours)
	if err != nil {
		handleInviteLinkError(c, err)
		return
	}
	Success(c, link)
}

// GetLinkInfo handles GET /invite-links/:code (public - no room context needed)
func (h *InviteLinkHandler) GetLinkInfo(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, 40001, "閭€璇风爜涓嶈兘涓虹┖")
		return
	}

	link, room, err := h.svc.GetLinkInfo(c.Request.Context(), code)
	if err != nil {
		handleInviteLinkError(c, err)
		return
	}
	Success(c, gin.H{"link": link, "room": room})
}

// UseLink handles POST /invite-links/:code/use
func (h *InviteLinkHandler) UseLink(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		Error(c, http.StatusBadRequest, 40001, "閭€璇风爜涓嶈兘涓虹┖")
		return
	}

	userID := c.GetInt64("user_id")
	result, err := h.svc.UseLink(c.Request.Context(), userID, code)
	if err != nil {
		handleInviteLinkError(c, err)
		return
	}
	Success(c, result)
}

// ListLinks handles GET /rooms/:id/invite-links
func (h *InviteLinkHandler) ListLinks(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	links, err := h.svc.ListLinks(c.Request.Context(), userID, roomID)
	if err != nil {
		handleInviteLinkError(c, err)
		return
	}
	if links == nil {
		links = []*model.InviteLink{}
	}
	Success(c, links)
}

// DeleteLink handles DELETE /rooms/:id/invite-links/:link_id
func (h *InviteLinkHandler) DeleteLink(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}
	linkID, ok := parseIDParam(c, "link_id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.svc.DeleteLink(c.Request.Context(), userID, roomID, linkID, c.GetString("role")); err != nil {
		handleInviteLinkError(c, err)
		return
	}
	Success(c, nil)
}

func handleInviteLinkError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInviteLinkNotFound):
		Error(c, http.StatusNotFound, 40401, "閭€璇烽摼鎺ヤ笉瀛樺湪鎴栧凡澶辨晥")
	case errors.Is(err, service.ErrInviteLinkExpired):
		Error(c, http.StatusGone, 41001, "閭€璇烽摼鎺ュ凡杩囨湡")
	case errors.Is(err, service.ErrInviteLinkMaxUses):
		Error(c, http.StatusGone, 41002, "资源已失效")
	case errors.Is(err, service.ErrInviteLinkForbid):
		Error(c, http.StatusForbidden, 40301, "没有权限")
	case errors.Is(err, service.ErrNotRoomMember):
		Error(c, http.StatusForbidden, 40301, "涓嶆槸鎴块棿鎴愬憳")
	case errors.Is(err, service.ErrRoomNotFound):
		Error(c, http.StatusNotFound, 40401, "资源不存在")
	case errors.Is(err, service.ErrAlreadyMember):
		Error(c, http.StatusConflict, 40902, "资源冲突")
	default:
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
	}
}
