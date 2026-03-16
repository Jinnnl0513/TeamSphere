package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

// AdminHandler handles all /admin routes.
// All endpoints require system admin+ role (enforced by middleware).
type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// ListUsers handles GET /admin/users?page=1&page_size=20.
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	users, total, err := h.adminService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}

	type userInfo struct {
		ID           int64  `json:"id"`
		Username     string `json:"username"`
		AvatarURL    string `json:"avatar_url"`
		Bio          string `json:"bio"`
		ProfileColor string `json:"profile_color"`
		Role         string `json:"role"`
		DeletedAt    any    `json:"deleted_at,omitempty"`
		CreatedAt    any    `json:"created_at"`
	}

	items := make([]userInfo, 0, len(users))
	for _, u := range users {
		items = append(items, userInfo{
			ID:           u.ID,
			Username:     u.Username,
			AvatarURL:    u.AvatarURL,
			Bio:          u.Bio,
			ProfileColor: u.ProfileColor,
			Role:         u.Role,
			DeletedAt:    u.DeletedAt,
			CreatedAt:    u.CreatedAt,
		})
	}

	Success(c, gin.H{
		"users":     items,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

type updateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

// UpdateUserRole handles PUT /admin/users/:id/role.
func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	targetID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}

	actorID := c.GetInt64("user_id")
	actorRole := c.GetString("role")

	err := h.adminService.UpdateUserRole(c.Request.Context(), actorID, targetID, actorRole, req.Role)
	if err != nil {
		handleAdminError(c, err)
		return
	}
	Success(c, nil)
}

// DeleteUser handles DELETE /admin/users/:id.
func (h *AdminHandler) DeleteUser(c *gin.Context) {
	targetID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	actorID := c.GetInt64("user_id")

	err := h.adminService.DeleteUser(c.Request.Context(), actorID, targetID)
	if err != nil {
		handleAdminError(c, err)
		return
	}
	Success(c, nil)
}

// ListRooms handles GET /admin/rooms?page=1&page_size=20.
func (h *AdminHandler) ListRooms(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	rooms, total, err := h.adminService.ListRooms(c.Request.Context(), page, pageSize)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}

	Success(c, gin.H{
		"rooms":     rooms,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// DeleteRoom handles DELETE /admin/rooms/:id.
func (h *AdminHandler) DeleteRoom(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	err := h.adminService.DeleteRoom(c.Request.Context(), roomID)
	if err != nil {
		handleAdminError(c, err)
		return
	}
	Success(c, nil)
}

// GetStats handles GET /admin/stats.
func (h *AdminHandler) GetStats(c *gin.Context) {
	stats, err := h.adminService.GetStats(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, stats)
}

// GetSettings handles GET /admin/settings.
func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings, err := h.adminService.GetSettings(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, settings)
}

type updateSettingsRequest struct {
	Settings map[string]string `json:"settings" binding:"required"`
}

// UpdateSettings handles PUT /admin/settings.
func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}

	updates := req.Settings
	if updates == nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	for k, v := range updates {
		if v == "********" && strings.Contains(k, "client_secret") {
			delete(updates, k)
		}
	}

	err := h.adminService.UpdateSettings(c.Request.Context(), updates)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, nil)
}

// GetEmailSettings handles GET /admin/email.
func (h *AdminHandler) GetEmailSettings(c *gin.Context) {
	settings, err := h.adminService.GetEmailSettings(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	resp := gin.H{
		"enabled":      settings.Enabled,
		"smtp_host":    settings.SMTPHost,
		"smtp_port":    settings.SMTPPort,
		"username":     settings.Username,
		"password":     "********",
		"from_address": settings.FromAddress,
		"from_name":    settings.FromName,
	}
	Success(c, resp)
}

type updateEmailRequest struct {
	Enabled     bool   `json:"enabled"`
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    int    `json:"smtp_port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
}

// UpdateEmailSettings handles PUT /admin/email.
func (h *AdminHandler) UpdateEmailSettings(c *gin.Context) {
	var req updateEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}

	settings := &service.EmailSettings{
		Enabled:     req.Enabled,
		SMTPHost:    req.SMTPHost,
		SMTPPort:    req.SMTPPort,
		Username:    req.Username,
		Password:    req.Password,
		FromAddress: req.FromAddress,
		FromName:    req.FromName,
	}

	if req.Password == "********" {
		existing, err := h.adminService.GetEmailSettings(c.Request.Context())
		if err == nil && existing != nil {
			settings.Password = existing.Password
		}
	}

	err := h.adminService.UpdateEmailSettings(c.Request.Context(), settings)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, nil)
}

type announcementRequest struct {
	Content string `json:"content"`
}

// SetAnnouncement handles POST /admin/announcement.
func (h *AdminHandler) SetAnnouncement(c *gin.Context) {
	var req announcementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}

	err := h.adminService.SetAnnouncement(c.Request.Context(), req.Content)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, nil)
}

// GetAnnouncement handles GET /admin/announcement.
func (h *AdminHandler) GetAnnouncement(c *gin.Context) {
	content, err := h.adminService.GetAnnouncement(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, gin.H{"content": content})
}

func handleAdminError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUserNotFound):
		Error(c, http.StatusNotFound, 40401, "user not found")
	case errors.Is(err, service.ErrRoomNotFound):
		Error(c, http.StatusNotFound, 40401, "room not found")
	case errors.Is(err, service.ErrNoPermission):
		Error(c, http.StatusForbidden, 40301, "no permission")
	case errors.Is(err, service.ErrAdminSelfRole):
		Error(c, http.StatusBadRequest, 40001, "cannot change your own role")
	case errors.Is(err, service.ErrAdminSelfDelete):
		Error(c, http.StatusBadRequest, 40001, "cannot delete your own account")
	case errors.Is(err, service.ErrAdminInvalidRole):
		Error(c, http.StatusBadRequest, 40001, "invalid role")
	case errors.Is(err, service.ErrAdminDeleteOwner):
		Error(c, http.StatusForbidden, 40301, "cannot modify the system owner")
	case errors.Is(err, service.ErrOwnsRooms):
		Error(c, http.StatusConflict, 40901, "target user still owns one or more rooms")
	default:
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
	}
}
