package handler

import (
	"net/http"
	"strconv"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

// List handles GET /notifications.
func (h *NotificationHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	unreadOnly := c.Query("unread") == "true"
	limit := parseOptionalInt(c.Query("limit"), 20)
	items, err := h.service.List(c.Request.Context(), userID, unreadOnly, limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, items)
}

// MarkRead handles PUT /notifications/:id/read.
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		Error(c, http.StatusBadRequest, 40001, "invalid notification id")
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.service.MarkRead(c.Request.Context(), userID, id); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, nil)
}
