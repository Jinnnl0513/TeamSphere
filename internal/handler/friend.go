package handler

import (
	"errors"
	"net/http"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

type FriendHandler struct {
	friendService *service.FriendService
}

func NewFriendHandler(friendService *service.FriendService) *FriendHandler {
	return &FriendHandler{friendService: friendService}
}

// ListFriends handles GET /friends.
func (h *FriendHandler) ListFriends(c *gin.Context) {
	userID := c.GetInt64("user_id")
	friends, err := h.friendService.ListFriends(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	Success(c, friends)
}

type sendRequestBody struct {
	Username string `json:"username" binding:"required"`
}

// SendRequest handles POST /friends/request.
func (h *FriendHandler) SendRequest(c *gin.Context) {
	var req sendRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
		return
	}

	userID := c.GetInt64("user_id")

	// Look up target user by username
	target, err := h.friendService.SearchUsers(c.Request.Context(), req.Username, userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}

	// Find exact match
	var targetUserID int64
	for _, u := range target {
		if u.Username == req.Username {
			targetUserID = u.ID
			break
		}
	}
	if targetUserID == 0 {
		Error(c, http.StatusNotFound, 40401, "资源不存在")
		return
	}

	friendship, err := h.friendService.SendRequest(c.Request.Context(), userID, targetUserID)
	if err != nil {
		handleFriendError(c, err)
		return
	}
	Success(c, friendship)
}

// ListPendingRequests handles GET /friends/requests.
func (h *FriendHandler) ListPendingRequests(c *gin.Context) {
	userID := c.GetInt64("user_id")
	requests, err := h.friendService.ListPendingRequests(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	Success(c, requests)
}

type respondRequestBody struct {
	Action string `json:"action" binding:"required"` // "accept" or "reject"
}

// RespondRequest handles PUT /friends/requests/:id.
func (h *FriendHandler) RespondRequest(c *gin.Context) {
	requestID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var req respondRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "鎿嶄綔绫诲瀷(action)蹇呭～锛屼笖鍙兘鏄?'accept' 鎴?'reject'")
		return
	}
	if req.Action != "accept" && req.Action != "reject" {
		Error(c, http.StatusBadRequest, 40001, "鎿嶄綔绫诲瀷(action)蹇呴』鏄?'accept' 鎴?'reject'")
		return
	}

	userID := c.GetInt64("user_id")

	if req.Action == "accept" {
		if err := h.friendService.AcceptRequest(c.Request.Context(), userID, requestID); err != nil {
			handleFriendError(c, err)
			return
		}
	} else {
		if err := h.friendService.RejectRequest(c.Request.Context(), userID, requestID); err != nil {
			handleFriendError(c, err)
			return
		}
	}
	Success(c, nil)
}

// DeleteFriend handles DELETE /friends/:id.
func (h *FriendHandler) DeleteFriend(c *gin.Context) {
	friendshipID, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	userID := c.GetInt64("user_id")
	if err := h.friendService.DeleteFriend(c.Request.Context(), userID, friendshipID); err != nil {
		handleFriendError(c, err)
		return
	}
	Success(c, nil)
}

// SearchUsers handles GET /users/search?q=xxx.
func (h *FriendHandler) SearchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		Error(c, http.StatusBadRequest, 40001, "璇锋眰鍙傛暟 q 涓嶈兘涓虹┖")
		return
	}

	userID := c.GetInt64("user_id")
	users, err := h.friendService.SearchUsers(c.Request.Context(), query, userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
		return
	}
	Success(c, users)
}

func handleFriendError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFriendSelf):
		Error(c, http.StatusBadRequest, 40001, "请求参数错误")
	case errors.Is(err, service.ErrFriendExists):
		Error(c, http.StatusConflict, 40901, "资源冲突")
	case errors.Is(err, service.ErrFriendReqPending):
		Error(c, http.StatusConflict, 40903, "宸叉湁寰呭鐞嗙殑濂藉弸璇锋眰")
	case errors.Is(err, service.ErrFriendReqNotFound):
		Error(c, http.StatusNotFound, 40401, "资源不存在")
	case errors.Is(err, service.ErrNotFriends):
		Error(c, http.StatusNotFound, 40401, "资源不存在")
	case errors.Is(err, service.ErrNoPermission):
		Error(c, http.StatusForbidden, 40301, "娌℃湁鏉冮檺")
	default:
		Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
	}
}
