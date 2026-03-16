package handler

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService   *service.UserService
	uploadService *service.UploadService
}

func NewUserHandler(userService *service.UserService, uploadService *service.UploadService) *UserHandler {
	return &UserHandler{userService: userService, uploadService: uploadService}
}

// GetMe handles GET /users/me.
func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetInt64("user_id")
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			Error(c, http.StatusNotFound, 40401, "user not found")
			return
		}
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, gin.H{
		"id":                user.ID,
		"username":          user.Username,
		"avatar_url":        user.AvatarURL,
		"bio":               user.Bio,
		"profile_color":     user.ProfileColor,
		"role":              user.Role,
		"email":             user.Email,
		"email_verified_at": user.EmailVerifiedAt,
	})
}

// GetByID handles GET /users/:id.
func (h *UserHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	var userID int64
	if _, err := fmt.Sscanf(idStr, "%d", &userID); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid user id")
		return
	}
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			Error(c, http.StatusNotFound, 40401, "user not found")
			return
		}
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, user.ToInfo())
}

var profileColorRegex = regexp.MustCompile(`^#[0-9a-fA-F]{3,6}$`)

type updateBioRequest struct {
	Bio          string `json:"bio" binding:"max=300"`
	ProfileColor string `json:"profile_color" binding:"max=7"`
}

// UpdateProfile handles PUT /users/me/profile.
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req updateBioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	if req.ProfileColor != "" && !profileColorRegex.MatchString(req.ProfileColor) {
		Error(c, http.StatusBadRequest, 40001, "invalid profile color")
		return
	}

	userID := c.GetInt64("user_id")
	user, err := h.userService.UpdateBioAndColor(c.Request.Context(), userID, req.Bio, req.ProfileColor)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, user.ToInfo())
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword handles PUT /users/me/password.
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	if !service.ValidatePassword(req.NewPassword) {
		Error(c, http.StatusBadRequest, 40001, "new password must be 8-128 characters with upper/lowercase and number")
		return
	}

	userID := c.GetInt64("user_id")
	claims := c.MustGet("claims").(*service.AuthClaims)
	err := h.userService.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword, claims)
	if err != nil {
		if errors.Is(err, service.ErrWrongPassword) {
			Error(c, http.StatusBadRequest, 40001, "wrong password")
			return
		}
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, nil)
}

// UploadAvatar handles POST /users/me/avatar.
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Error(c, http.StatusBadRequest, 40001, "file is required")
		return
	}

	result, err := h.uploadService.SaveFile(file)
	if err != nil {
		if errors.Is(err, service.ErrFileTooLarge) || errors.Is(err, service.ErrFileTypeNotAllowed) {
			Error(c, http.StatusBadRequest, 40001, err.Error())
		} else {
			Error(c, http.StatusInternalServerError, 50001, "file upload failed")
		}
		return
	}

	userID := c.GetInt64("user_id")
	user, err := h.userService.UpdateAvatar(c.Request.Context(), userID, result.URL)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, user.ToInfo())
}

type deleteAccountRequest struct {
	Password string `json:"password" binding:"required"`
}

// DeleteAccount handles DELETE /users/me.
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	var req deleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "password is required")
		return
	}

	userID := c.GetInt64("user_id")
	claims := c.MustGet("claims").(*service.AuthClaims)
	err := h.userService.DeleteAccount(c.Request.Context(), userID, req.Password, claims)
	if err != nil {
		if errors.Is(err, service.ErrWrongPassword) {
			Error(c, http.StatusBadRequest, 40001, "wrong password")
			return
		}
		if errors.Is(err, service.ErrOwnsRooms) {
			Error(c, http.StatusConflict, 40901, "transfer or delete owned rooms before deleting this account")
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			Error(c, http.StatusNotFound, 40401, "user not found")
			return
		}
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, nil)
}
