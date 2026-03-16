package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

var usernameRegex = regexp.MustCompile(`^\w{3,32}$`)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

const passwordResetSentMessage = "If the email is registered, a reset code has been sent"

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Username          string `json:"username" binding:"required"`
	Password          string `json:"password" binding:"required"`
	Email             string `json:"email"`
	VerificationToken string `json:"verification_token"`
}

type loginRequest struct {
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	TOTPCode     string `json:"totp_code"`
	RecoveryCode string `json:"recovery_code"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type sendCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

type verifyEmailRequest struct {
	Email string `json:"email" binding:"required"`
	Code  string `json:"code" binding:"required"`
}

type sendPasswordResetCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

type resetPasswordRequest struct {
	Email       string `json:"email" binding:"required"`
	Code        string `json:"code" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type totpVerifyRequest struct {
	Code string `json:"code" binding:"required"`
}

type totpVerifyLoginRequest struct {
	Challenge    string `json:"challenge" binding:"required"`
	Code         string `json:"code"`
	RecoveryCode string `json:"recovery_code"`
}

type totpSetupRequiredRequest struct {
	SetupToken string `json:"setup_token" binding:"required"`
}

type totpEnableRequiredRequest struct {
	SetupToken string `json:"setup_token" binding:"required"`
	Code       string `json:"code" binding:"required"`
}

type revokeSessionRequest struct {
	SessionID int64 `json:"session_id" binding:"required"`
}

type revokeOtherSessionsRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if !usernameRegex.MatchString(req.Username) {
		Error(c, http.StatusBadRequest, 40001, "username must be 3-32 characters using letters, numbers, or underscore")
		return
	}
	if !service.ValidatePassword(req.Password) {
		Error(c, http.StatusBadRequest, 40001, "password must be 8-128 characters with upper/lowercase and number")
		return
	}
	if req.Email != "" && !emailRegex.MatchString(req.Email) {
		Error(c, http.StatusBadRequest, 40001, "invalid email address")
		return
	}

	userAgent := c.Request.UserAgent()
	result, err := h.authService.Register(
		c.Request.Context(),
		req.Username,
		req.Password,
		req.Email,
		req.VerificationToken,
		strPtr(c.ClientIP()),
		strPtr(userAgent),
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRegisterDisabled):
			Error(c, http.StatusForbidden, 40301, "registration is disabled")
		case errors.Is(err, service.ErrUsernameTaken):
			Error(c, http.StatusConflict, 40901, "username already exists")
		case errors.Is(err, service.ErrEmailTaken):
			Error(c, http.StatusConflict, 40901, "email already registered")
		case errors.Is(err, service.ErrEmailNotVerified):
			Error(c, http.StatusBadRequest, 40001, "email verification is required")
		case errors.Is(err, service.ErrInvalidVerifyToken):
			Error(c, http.StatusBadRequest, 40001, "verification token is invalid or expired")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}

	Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          result.User.ToInfo(),
	})
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	req.Username = strings.TrimSpace(req.Username)

	userAgent := c.Request.UserAgent()
	result, err := h.authService.Login(c.Request.Context(), req.Username, req.Password, req.TOTPCode, req.RecoveryCode, c.ClientIP(), strPtr(userAgent))
	if err != nil {
		var setupErr *service.TOTPSetupRequiredError
		if errors.As(err, &setupErr) {
			ErrorWithData(c, http.StatusUnauthorized, 40106, "totp setup required", gin.H{
				"setup_token": setupErr.Challenge,
			})
			return
		}
		if errors.Is(err, service.ErrLoginLocked) {
			Error(c, http.StatusTooManyRequests, 42901, "too many failed attempts, try again later")
			return
		}
		if errors.Is(err, service.ErrInvalidCredentials) {
			Error(c, http.StatusUnauthorized, 40101, "invalid username or password")
			return
		}
		if errors.Is(err, service.ErrTOTPRequired) {
			Error(c, http.StatusUnauthorized, 40103, "totp required")
			return
		}
		if errors.Is(err, service.ErrTOTPInvalid) {
			Error(c, http.StatusUnauthorized, 40104, "totp invalid")
			return
		}
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}

	Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          result.User.ToInfo(),
	})
}

// Logout handles POST /auth/logout.
func (h *AuthHandler) Logout(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		Error(c, http.StatusUnauthorized, 40101, "authentication required")
		return
	}

	var req logoutRequest
	_ = c.ShouldBindJSON(&req)

	authClaims, ok := claims.(*service.AuthClaims)
	if !ok {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	if err := h.authService.Logout(c.Request.Context(), authClaims, req.RefreshToken); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}

	Success(c, nil)
}

// RefreshToken handles POST /auth/refresh.
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}

	result, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRefreshTokenInvalid):
			Error(c, http.StatusUnauthorized, 40101, "refresh token is invalid or expired")
		case errors.Is(err, service.ErrRefreshTokenRevoked):
			Error(c, http.StatusUnauthorized, 40101, "refresh token has been revoked")
		default:
			Error(c, http.StatusUnauthorized, 40101, "refresh token validation failed")
		}
		return
	}

	Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          result.User.ToInfo(),
	})
}

// SendCode handles POST /auth/send-code.
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req sendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}

	if !emailRegex.MatchString(req.Email) {
		Error(c, http.StatusBadRequest, 40001, "invalid email address")
		return
	}

	err := h.authService.SendVerificationCode(c.Request.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailDisabled):
			Error(c, http.StatusBadRequest, 40001, "email is disabled; enable it in admin settings first")
		case errors.Is(err, service.ErrEmailTaken):
			Error(c, http.StatusConflict, 40901, "email already registered")
		case errors.Is(err, service.ErrCodeTooFrequent):
			Error(c, http.StatusTooManyRequests, 42901, "please wait 60 seconds before requesting another code")
		default:
			slog.Error("failed to send verification code", "error", err)
			Error(c, http.StatusInternalServerError, 50001, "failed to send verification code")
		}
		return
	}

	Success(c, gin.H{"message": "verification code sent"})
}

// VerifyEmail handles POST /auth/verify-email.
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}

	token, err := h.authService.VerifyEmailCode(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCodeNotFound):
			Error(c, http.StatusBadRequest, 40001, "no valid verification code found")
		case errors.Is(err, service.ErrCodeInvalid):
			Error(c, http.StatusBadRequest, 40001, "invalid verification code")
		case errors.Is(err, service.ErrCodeTooManyRetries):
			Error(c, http.StatusTooManyRequests, 42901, "too many attempts; request a new code")
		default:
			Error(c, http.StatusInternalServerError, 50001, "email verification failed")
		}
		return
	}

	Success(c, gin.H{"verification_token": token})
}

// SendPasswordResetCode handles POST /auth/password/reset-code.
func (h *AuthHandler) SendPasswordResetCode(c *gin.Context) {
	var req sendPasswordResetCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	if !emailRegex.MatchString(req.Email) {
		Error(c, http.StatusBadRequest, 40001, "invalid email address")
		return
	}

	err := h.authService.SendPasswordResetCode(c.Request.Context(), req.Email)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailDisabled):
			Error(c, http.StatusBadRequest, 40001, "email is disabled; contact an administrator")
		case errors.Is(err, service.ErrUserNotFound):
			Success(c, gin.H{"message": passwordResetSentMessage})
			return
		case errors.Is(err, service.ErrCodeTooFrequent):
			Error(c, http.StatusTooManyRequests, 42901, "please wait 60 seconds before requesting another code")
		default:
			slog.Error("failed to send password reset code", "error", err)
			Error(c, http.StatusInternalServerError, 50001, "failed to send password reset code")
		}
		return
	}

	Success(c, gin.H{"message": passwordResetSentMessage})
}

// ResetPassword handles POST /auth/password/reset.
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Code = strings.TrimSpace(req.Code)

	if !emailRegex.MatchString(req.Email) {
		Error(c, http.StatusBadRequest, 40001, "invalid email address")
		return
	}
	if req.Code == "" {
		Error(c, http.StatusBadRequest, 40001, "verification code is required")
		return
	}
	if !service.ValidatePassword(req.NewPassword) {
		Error(c, http.StatusBadRequest, 40001, "new password must be 8-128 characters with upper/lowercase and number")
		return
	}

	err := h.authService.ResetPasswordByCode(c.Request.Context(), req.Email, req.Code, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserNotFound):
			Error(c, http.StatusNotFound, 40401, "user not found")
		case errors.Is(err, service.ErrCodeNotFound):
			Error(c, http.StatusBadRequest, 40001, "no valid verification code found")
		case errors.Is(err, service.ErrCodeInvalid):
			Error(c, http.StatusBadRequest, 40001, "invalid verification code")
		case errors.Is(err, service.ErrCodeTooManyRetries):
			Error(c, http.StatusTooManyRequests, 42901, "too many attempts; request a new code")
		case errors.Is(err, service.ErrWeakPassword):
			Error(c, http.StatusBadRequest, 40001, "password does not meet complexity requirements")
		default:
			slog.Error("failed to reset password", "error", err)
			Error(c, http.StatusInternalServerError, 50001, "failed to reset password")
		}
		return
	}

	Success(c, gin.H{"message": "password reset successful"})
}

// EmailRequired handles GET /auth/email-required.
func (h *AuthHandler) EmailRequired(c *gin.Context) {
	ctx := c.Request.Context()
	required := h.authService.IsEmailRequired(ctx)
	allowRegister := h.authService.IsRegisterAllowed(ctx)
	enabled := h.authService.IsEmailEnabled(ctx)
	Success(c, gin.H{
		"allow_register":       allowRegister,
		"email_required":       required,
		"email_enabled":        enabled,
	})
}

// TOTPStatus handles GET /auth/2fa/status.
func (h *AuthHandler) TOTPStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")
	enabled, err := h.authService.GetTOTPStatus(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	role := c.GetString("role")
	policy := h.authService.Get2FAPolicy(c.Request.Context())
	required := h.authService.Is2FARequiredForRole(c.Request.Context(), role)
	Success(c, gin.H{
		"enabled":  enabled,
		"policy":   policy,
		"required": required,
	})
}

// TOTPSetup handles POST /auth/2fa/setup.
func (h *AuthHandler) TOTPSetup(c *gin.Context) {
	userID := c.GetInt64("user_id")
	result, err := h.authService.StartTOTPSetup(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, result)
}

// TOTPEnable handles POST /auth/2fa/enable.
func (h *AuthHandler) TOTPEnable(c *gin.Context) {
	var req totpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userID := c.GetInt64("user_id")
	codes, err := h.authService.EnableTOTP(c.Request.Context(), userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPInvalid):
			Error(c, http.StatusBadRequest, 40001, "totp invalid")
		case errors.Is(err, service.ErrTOTPNotSetup):
			Error(c, http.StatusBadRequest, 40001, "totp not setup")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}
	Success(c, gin.H{"recovery_codes": codes})
}

// TOTPDisable handles POST /auth/2fa/disable.
func (h *AuthHandler) TOTPDisable(c *gin.Context) {
	var req totpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.authService.DisableTOTP(c.Request.Context(), userID, req.Code); err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPPolicyRequired):
			Error(c, http.StatusForbidden, 40302, "2fa required by policy")
		case errors.Is(err, service.ErrUserNotFound):
			Error(c, http.StatusNotFound, 40401, "user not found")
		case errors.Is(err, service.ErrTOTPInvalid):
			Error(c, http.StatusBadRequest, 40001, "totp invalid")
		case errors.Is(err, service.ErrTOTPNotSetup):
			Error(c, http.StatusBadRequest, 40001, "totp not setup")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}
	Success(c, nil)
}

// TOTPRecoveryStatus handles GET /auth/2fa/recovery-codes/status.
func (h *AuthHandler) TOTPRecoveryStatus(c *gin.Context) {
	userID := c.GetInt64("user_id")
	count, err := h.authService.RecoveryCodesStatus(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, gin.H{"remaining": count})
}

// TOTPRecoveryRegen handles POST /auth/2fa/recovery-codes/regen.
func (h *AuthHandler) TOTPRecoveryRegen(c *gin.Context) {
	var req totpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userID := c.GetInt64("user_id")
	codes, err := h.authService.RegenRecoveryCodes(c.Request.Context(), userID, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPInvalid):
			Error(c, http.StatusBadRequest, 40001, "totp invalid")
		case errors.Is(err, service.ErrTOTPNotSetup):
			Error(c, http.StatusBadRequest, 40001, "totp not setup")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}
	Success(c, gin.H{"recovery_codes": codes})
}

// TOTPVerifyLogin handles POST /auth/2fa/verify-login.
func (h *AuthHandler) TOTPVerifyLogin(c *gin.Context) {
	var req totpVerifyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userAgent := c.Request.UserAgent()
	result, err := h.authService.VerifyTwoFALogin(c.Request.Context(), req.Challenge, req.Code, req.RecoveryCode, strPtr(c.ClientIP()), strPtr(userAgent))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPChallengeInvalid):
			Error(c, http.StatusUnauthorized, 40105, "totp challenge invalid")
		case errors.Is(err, service.ErrTOTPPolicyRequired):
			Error(c, http.StatusForbidden, 40302, "2fa required by policy")
		case errors.Is(err, service.ErrTOTPRequired):
			Error(c, http.StatusUnauthorized, 40103, "totp required")
		case errors.Is(err, service.ErrTOTPInvalid):
			Error(c, http.StatusUnauthorized, 40104, "totp invalid")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}
	Success(c, gin.H{
		"token":         result.Token,
		"refresh_token": result.RefreshToken,
		"user":          result.User.ToInfo(),
	})
}

// TOTPSetupRequired handles POST /auth/2fa/setup-required.
func (h *AuthHandler) TOTPSetupRequired(c *gin.Context) {
	var req totpSetupRequiredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	result, _, err := h.authService.StartTOTPSetupWithChallenge(c.Request.Context(), req.SetupToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPChallengeInvalid):
			Error(c, http.StatusUnauthorized, 40105, "totp challenge invalid")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}
	Success(c, result)
}

// TOTPEnableRequired handles POST /auth/2fa/enable-required.
func (h *AuthHandler) TOTPEnableRequired(c *gin.Context) {
	var req totpEnableRequiredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userAgent := c.Request.UserAgent()
	codes, userID, err := h.authService.EnableTOTPWithChallenge(c.Request.Context(), req.SetupToken, req.Code)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTOTPChallengeInvalid):
			Error(c, http.StatusUnauthorized, 40105, "totp challenge invalid")
		case errors.Is(err, service.ErrTOTPInvalid):
			Error(c, http.StatusBadRequest, 40001, "totp invalid")
		case errors.Is(err, service.ErrTOTPNotSetup):
			Error(c, http.StatusBadRequest, 40001, "totp not setup")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil || user == nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	accessToken, err := h.authService.GenerateAccessTokenForUser(user)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	refreshToken, err := h.authService.GenerateRefreshTokenForUser(c.Request.Context(), user.ID, strPtr(c.ClientIP()), strPtr(userAgent), nil)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, gin.H{
		"token":          accessToken,
		"refresh_token":  refreshToken,
		"user":           user.ToInfo(),
		"recovery_codes": codes,
	})
}

// ListSessions handles GET /auth/sessions.
func (h *AuthHandler) ListSessions(c *gin.Context) {
	userID := c.GetInt64("user_id")
	sessions, err := h.authService.ListSessions(c.Request.Context(), userID)
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	currentHash := ""
	if token := strings.TrimSpace(c.GetHeader("X-Refresh-Token")); token != "" {
		hash := sha256.Sum256([]byte(token))
		currentHash = hex.EncodeToString(hash[:])
	}

	type sessionInfo struct {
		ID         int64   `json:"id"`
		CreatedAt  any     `json:"created_at"`
		LastUsedAt any     `json:"last_used_at,omitempty"`
		ExpiresAt  any     `json:"expires_at"`
		RevokedAt  any     `json:"revoked_at,omitempty"`
		IPAddress  *string `json:"ip_address,omitempty"`
		UserAgent  *string `json:"user_agent,omitempty"`
		DeviceName *string `json:"device_name,omitempty"`
		IsCurrent  bool    `json:"is_current"`
	}

	items := make([]sessionInfo, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, sessionInfo{
			ID:         s.ID,
			CreatedAt:  s.CreatedAt,
			LastUsedAt: s.LastUsedAt,
			ExpiresAt:  s.ExpiresAt,
			RevokedAt:  s.RevokedAt,
			IPAddress:  s.IPAddress,
			UserAgent:  s.UserAgent,
			DeviceName: s.DeviceName,
			IsCurrent:  currentHash != "" && s.TokenHash == currentHash,
		})
	}

	Success(c, gin.H{"sessions": items})
}

// RevokeSession handles POST /auth/sessions/revoke.
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	var req revokeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.authService.RevokeSession(c.Request.Context(), userID, req.SessionID); err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, nil)
}

// RevokeOtherSessions handles POST /auth/sessions/revoke-others.
func (h *AuthHandler) RevokeOtherSessions(c *gin.Context) {
	var req revokeOtherSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid request payload")
		return
	}
	userID := c.GetInt64("user_id")
	if err := h.authService.RevokeOtherSessions(c.Request.Context(), userID, req.RefreshToken); err != nil {
		Error(c, http.StatusBadRequest, 40001, "invalid refresh token")
		return
	}
	Success(c, nil)
}

// OAuthProviders handles GET /auth/oauth/providers.
func (h *AuthHandler) OAuthProviders(c *gin.Context) {
	providers, err := h.authService.ListOAuthProviders(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, 50001, "internal server error")
		return
	}
	Success(c, gin.H{"providers": providers})
}

// OAuthStart handles GET /auth/oauth/:provider/start.
func (h *AuthHandler) OAuthStart(c *gin.Context) {
	provider := c.Param("provider")
	redirect := sanitizeOAuthRedirect(c.Query("redirect"))
	authURL, err := h.authService.StartOAuth(c.Request.Context(), provider, redirect)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOAuthDisabled):
			Error(c, http.StatusForbidden, 40301, "oauth is disabled")
		case errors.Is(err, service.ErrOAuthProviderMissing):
			Error(c, http.StatusBadRequest, 40001, "oauth provider not configured")
		default:
			Error(c, http.StatusInternalServerError, 50001, "internal server error")
		}
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

// OAuthCallback handles GET /auth/oauth/:provider/callback.
func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	if errCode := strings.TrimSpace(c.Query("error")); errCode != "" {
		errDesc := strings.TrimSpace(c.Query("error_description"))
		if errDesc == "" {
			errDesc = errCode
		}
		c.Redirect(http.StatusFound, "/login?oauth_error="+url.QueryEscape(errDesc))
		return
	}
	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	if code == "" || state == "" {
		c.Redirect(http.StatusFound, "/login?oauth_error=missing_code_or_state")
		return
	}

	userAgent := c.Request.UserAgent()
	result, redirectPath, challenge, setupToken, err := h.authService.HandleOAuthCallback(c.Request.Context(), provider, code, state, strPtr(c.ClientIP()), strPtr(userAgent))
	if err != nil {
		errMsg := "oauth_failed"
		switch {
		case errors.Is(err, service.ErrOAuthDisabled):
			errMsg = "oauth_disabled"
		case errors.Is(err, service.ErrOAuthProviderMissing):
			errMsg = "provider_not_configured"
		case errors.Is(err, service.ErrOAuthInvalidState):
			errMsg = "invalid_state"
		case errors.Is(err, service.ErrOAuthExchangeFailed):
			errMsg = "exchange_failed"
		case errors.Is(err, service.ErrOAuthProfileFailed):
			errMsg = "profile_failed"
		}
		c.Redirect(http.StatusFound, "/login?oauth_error="+url.QueryEscape(errMsg))
		return
	}

	if setupToken != "" {
		callbackURL := buildOAuthSetupURL(h.authService.GetOAuthFrontendBaseURL(c.Request.Context()), setupToken, redirectPath)
		c.Redirect(http.StatusFound, callbackURL)
		return
	}

	if challenge != "" {
		callbackURL := buildOAuthChallengeURL(h.authService.GetOAuthFrontendBaseURL(c.Request.Context()), challenge, redirectPath)
		c.Redirect(http.StatusFound, callbackURL)
		return
	}

	callbackURL := buildOAuthCallbackURL(h.authService.GetOAuthFrontendBaseURL(c.Request.Context()), result.Token, result.RefreshToken, redirectPath)
	c.Redirect(http.StatusFound, callbackURL)
}

func sanitizeOAuthRedirect(raw string) string {
	if raw == "" {
		return "/chat"
	}
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return raw
	}
	return "/chat"
}

func buildOAuthCallbackURL(base, token, refreshToken, redirect string) string {
	path := "/oauth/callback"
	if base != "" {
		base = strings.TrimRight(base, "/")
		path = base + path
	}
	if redirect == "" {
		redirect = "/chat"
	}
	params := url.Values{}
	params.Set("token", token)
	params.Set("refresh_token", refreshToken)
	params.Set("redirect", redirect)
	return path + "?" + params.Encode()
}

func buildOAuthChallengeURL(base, challenge, redirect string) string {
	path := "/oauth/callback"
	if base != "" {
		base = strings.TrimRight(base, "/")
		path = base + path
	}
	if redirect == "" {
		redirect = "/chat"
	}
	params := url.Values{}
	params.Set("challenge", challenge)
	params.Set("redirect", redirect)
	return path + "?" + params.Encode()
}

func buildOAuthSetupURL(base, setupToken, redirect string) string {
	path := "/oauth/callback"
	if base != "" {
		base = strings.TrimRight(base, "/")
		path = base + path
	}
	if redirect == "" {
		redirect = "/chat"
	}
	params := url.Values{}
	params.Set("setup_token", setupToken)
	params.Set("redirect", redirect)
	return path + "?" + params.Encode()
}

func strPtr(v string) *string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return &v
}
