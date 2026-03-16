package handler

import (
	"net/http"

	"github.com/teamsphere/server/internal/model"
	"github.com/gin-gonic/gin"
)

// Room CRUD

type createRoomRequest struct {
	Name        string `json:"name" binding:"required,max=64"`
	Description string `json:"description" binding:"max=256"`
}

// Create handles POST /rooms.
func (h *RoomHandler) Create(c *gin.Context) {
	var req createRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	room, err := h.roomService.Create(c.Request.Context(), userID, req.Name, req.Description)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, room)
}

// GetByID handles GET /rooms/:id.
func (h *RoomHandler) GetByID(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	room, err := h.roomService.GetByID(c.Request.Context(), userID, roomID)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, room)
}

type updateRoomRequest struct {
	Name        string `json:"name" binding:"required,max=64"`
	Description string `json:"description" binding:"max=256"`
}

// Update handles PUT /rooms/:id.
func (h *RoomHandler) Update(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req updateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")
	room, err := h.roomService.Update(c.Request.Context(), userID, c.GetString("role"), roomID, req.Name, req.Description)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, room)
}

// Delete handles DELETE /rooms/:id.
func (h *RoomHandler) Delete(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.Delete(c.Request.Context(), userID, c.GetString("role"), roomID); err != nil {
		handleRoomError(c, err)
		return
	}
	h.audit(c, "room_delete", "room", roomID, nil)
	Success(c, nil)
}

// List handles GET /rooms.
func (h *RoomHandler) List(c *gin.Context) {
	userID := c.GetInt64("user_id")
	rooms, err := h.roomService.ListByUser(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	if rooms == nil {
		rooms = []*model.Room{}
	}
	Success(c, rooms)
}

// DiscoverAll handles GET /rooms/discover.
func (h *RoomHandler) DiscoverAll(c *gin.Context) {
	rooms, err := h.roomService.DiscoverAll(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	if rooms == nil {
		rooms = []*model.Room{}
	}
	Success(c, rooms)
}

// JoinRoom handles POST /rooms/:id/join.
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.JoinRoom(c.Request.Context(), userID, roomID); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}

// LeaveRoom handles POST /rooms/:id/leave.
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	roomID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.roomService.LeaveRoom(c.Request.Context(), userID, roomID); err != nil {
		handleRoomError(c, err)
		return
	}
	Success(c, nil)
}
